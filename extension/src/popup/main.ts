import "./style.css";

document.addEventListener("DOMContentLoaded", () => {
  const app = document.getElementById("app")!;
  app.innerHTML = `
    <div class="p-4 flex flex-col gap-2">
      <h1 class="text-base font-semibold text-gray-900 dark:text-white">Keeper</h1>
      <p class="text-sm text-gray-500 dark:text-gray-400">Your secret manager is ready.</p>
    </div>
  `;
});
