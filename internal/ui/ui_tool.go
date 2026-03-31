package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"

	"github.com/davidsbond/keeper/internal/server/password"
	"github.com/davidsbond/keeper/internal/server/service"
	"github.com/davidsbond/keeper/internal/server/token"
	statusview "github.com/davidsbond/keeper/internal/ui/view/status"
	toolview "github.com/davidsbond/keeper/internal/ui/view/tool"
)

type (
	// The ToolHandler type is responsible for serving web interface pages for the tools section.
	ToolHandler struct {
		accounts AccountService
		tools    ToolService
	}

	// The ToolService interface describes types that provide user tool implementations.
	ToolService interface {
		// Export should return all the specified user's data.
		Export(accountID uuid.UUID) (service.Export, error)
		// Import should import vault data for the specified user from the given reader, interpreting
		// it according to the provided ImportSource.
		Import(accountID uuid.UUID, source service.ImportSource, data io.Reader) (service.ImportResult, error)
	}
)

// NewToolHandler returns a new instance of the ToolHandler type that will serve tool UIs using the provided
// AccountService and ToolService implementations.
func NewToolHandler(accounts AccountService, tools ToolService) *ToolHandler {
	return &ToolHandler{accounts: accounts, tools: tools}
}

// Register HTTP endpoints onto the provided http.ServeMux.
func (h *ToolHandler) Register(mux *http.ServeMux) {
	mux.Handle("GET /tools", requireToken(h.Index))
	mux.Handle("GET /tools/export", requireToken(h.Export))
	mux.Handle("POST /tools/export", requireToken(h.ExportCallback))
	mux.Handle("GET /tools/import", requireToken(h.Import))
	mux.Handle("POST /tools/import", requireToken(h.ImportCallback))
	mux.Handle("GET /tools/password-generator", requireToken(h.PasswordGenerator))
	mux.Handle("POST /tools/password-generator", requireToken(h.PasswordGeneratorCallback))
}

// Index renders the tools index view.
func (h *ToolHandler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	render(ctx, w, http.StatusOK, toolview.Index, toolview.ViewModel{
		DisplayName: account.DisplayName,
	})
}

// Export renders the export confirmation view.
func (h *ToolHandler) Export(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	render(ctx, w, http.StatusOK, toolview.Export, toolview.ExportViewModel{
		DisplayName: account.DisplayName,
	})
}

// The ExportForm type represents the form values submitted when calling ToolHandler.ExportCallback.
type ExportForm struct {
	// The user's password, required to confirm the export.
	Password string `form:"password"`
}

// Validate the form.
func (f ExportForm) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Password, validation.Required),
	)
}

// ExportCallback handles an export request. On success, it responds with a downloadable JSON file containing
// all the user's stored data.
func (h *ToolHandler) ExportCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	form, err := decode[ExportForm](r)

	model := toolview.ExportViewModel{
		DisplayName: account.DisplayName,
	}

	var ve validation.Errors
	switch {
	case errors.As(err, &ve):
		model.Validation.Errors = validationErrors(ve)
		render(ctx, w, http.StatusUnprocessableEntity, toolview.Export, model)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	err = h.accounts.VerifyPassword(tkn.ID(), form.Password)
	switch {
	case errors.Is(err, service.ErrInvalidPassword):
		model.Error.Message = "Password is incorrect"
		render(ctx, w, http.StatusUnprocessableEntity, toolview.Export, model)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	export, err := h.tools.Export(tkn.ID())
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirectToLogin(w, r)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="keeper_%d.json"`, time.Now().Unix()))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// Import renders the import form view.
func (h *ToolHandler) Import(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	render(ctx, w, http.StatusOK, toolview.Import, toolview.ImportViewModel{
		DisplayName: account.DisplayName,
	})
}

// ImportCallback handles a multipart form submission to import vault data. On success it renders the import
// result view showing counts of imported records and any per-item errors.
func (h *ToolHandler) ImportCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	model := toolview.ImportViewModel{
		DisplayName: account.DisplayName,
	}

	if err = r.ParseMultipartForm(10 << 20); err != nil {
		model.Error.Message = "Failed to read the uploaded file"
		render(ctx, w, http.StatusUnprocessableEntity, toolview.Import, model)
		return
	}

	source, err := service.ParseImportSource(r.FormValue("source"))
	if err != nil {
		model.Validation.Errors = map[string]string{"source": "Invalid import source selected"}
		render(ctx, w, http.StatusUnprocessableEntity, toolview.Import, model)
		return
	}

	file, _, err := r.FormFile("file")
	switch {
	case errors.Is(err, http.ErrMissingFile):
		model.Validation.Errors = map[string]string{"file": "A file is required"}
		render(ctx, w, http.StatusUnprocessableEntity, toolview.Import, model)
		return
	case err != nil:
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}
	defer file.Close()

	result, err := h.tools.Import(tkn.ID(), source, file)
	switch {
	case errors.Is(err, service.ErrReauthenticate):
		redirectToLogin(w, r)
		return
	case err != nil:
		model.Error.Message = err.Error()
		render(ctx, w, http.StatusUnprocessableEntity, toolview.Import, model)
		return
	}

	render(ctx, w, http.StatusOK, toolview.ImportResult, toolview.ImportResultViewModel{
		DisplayName: account.DisplayName,
		Logins:      result.Logins,
		Notes:       result.Notes,
		Cards:       result.Cards,
		Errors:      result.Errors,
	})
}

// PasswordGenerator renders the password generator view with default options.
func (h *ToolHandler) PasswordGenerator(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	render(ctx, w, http.StatusOK, toolview.PasswordGenerator, toolview.PasswordGeneratorViewModel{
		DisplayName: account.DisplayName,
		Length:      16,
		Uppercase:   true,
		Lowercase:   true,
		Numbers:     true,
		Symbols:     true,
	})
}

// The PasswordGeneratorForm type represents the form values submitted when calling ToolHandler.PasswordGeneratorCallback.
type PasswordGeneratorForm struct {
	// The desired password length.
	Length int `form:"length"`
	// Whether to include uppercase letters.
	Uppercase bool `form:"uppercase"`
	// Whether to include lowercase letters.
	Lowercase bool `form:"lowercase"`
	// Whether to include numeric digits.
	Numbers bool `form:"numbers"`
	// Whether to include symbols.
	Symbols bool `form:"symbols"`
}

// Validate the form.
func (f PasswordGeneratorForm) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Length, validation.Required, validation.Min(1), validation.Max(128)),
	)
}

// PasswordGeneratorCallback handles a password generation request. On success, it renders the generator view
// with the generated password and its strength rating.
func (h *ToolHandler) PasswordGeneratorCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tkn := token.FromContext(ctx)

	account, err := h.accounts.Get(tkn.ID())
	if err != nil {
		render(ctx, w, http.StatusInternalServerError, statusview.InternalServerError, statusview.InternalServerErrorViewModel{
			Detail: err.Error(),
		})
		return
	}

	form, err := decode[PasswordGeneratorForm](r)
	model := toolview.PasswordGeneratorViewModel{
		DisplayName: account.DisplayName,
		Length:      form.Length,
		Uppercase:   form.Uppercase,
		Lowercase:   form.Lowercase,
		Numbers:     form.Numbers,
		Symbols:     form.Symbols,
	}

	if model.Length == 0 {
		model.Length = 16
	}

	if _, ok := errors.AsType[validation.Errors](err); ok {
		model.Error.Message = "Please enter a valid length between 1 and 128."
		render(ctx, w, http.StatusUnprocessableEntity, toolview.PasswordGenerator, model)
		return
	}

	generated, err := password.Generate(password.GenerateOptions{
		Length:    form.Length,
		Uppercase: form.Uppercase,
		Lowercase: form.Lowercase,
		Numbers:   form.Numbers,
		Symbols:   form.Symbols,
	})
	if err != nil {
		model.Error.Message = err.Error()
		render(ctx, w, http.StatusUnprocessableEntity, toolview.PasswordGenerator, model)
		return
	}

	model.Password = generated
	model.Rating = password.Rate(generated)
	render(ctx, w, http.StatusOK, toolview.PasswordGenerator, model)
}
