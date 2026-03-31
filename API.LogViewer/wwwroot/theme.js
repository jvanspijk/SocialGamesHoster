window.themeManager = {
    storageKey: "api-logviewer-theme",
    cookieKey: "api-logviewer-theme",

    normalizeTheme: (theme) => theme === "dark" ? "dark" : "light",

    readCookieTheme: () => {
        const cookieMatch = document.cookie.match(/(?:^|; )api-logviewer-theme=([^;]+)/);
        if (!cookieMatch) {
            return null;
        }

        const parsed = decodeURIComponent(cookieMatch[1]);
        return parsed === "dark" || parsed === "light" ? parsed : null;
    },

    writeCookieTheme: (theme) => {
        const normalizedTheme = window.themeManager.normalizeTheme(theme);
        const oneYearInSeconds = 60 * 60 * 24 * 365;
        document.cookie = `${window.themeManager.cookieKey}=${encodeURIComponent(normalizedTheme)}; path=/; max-age=${oneYearInSeconds}; samesite=lax`;
    },

    getSavedTheme: () => {
        const storedTheme = localStorage.getItem(window.themeManager.storageKey);
        if (storedTheme === "dark" || storedTheme === "light") {
            return storedTheme;
        }

        const cookieTheme = window.themeManager.readCookieTheme();
        if (cookieTheme) {
            return cookieTheme;
        }

        const serverTheme = document.documentElement.getAttribute("data-theme");
        if (serverTheme === "dark" || serverTheme === "light") {
            return serverTheme;
        }

        return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
    },

    applyTheme: (theme) => {
        const normalizedTheme = window.themeManager.normalizeTheme(theme);
        document.documentElement.setAttribute("data-theme", normalizedTheme);
        document.documentElement.setAttribute("data-bs-theme", normalizedTheme);
        window.themeManager.syncToggleUi(normalizedTheme);
    },

    setTheme: (theme) => {
        const normalizedTheme = window.themeManager.normalizeTheme(theme);
        localStorage.setItem(window.themeManager.storageKey, normalizedTheme);
        window.themeManager.writeCookieTheme(normalizedTheme);
        window.themeManager.applyTheme(normalizedTheme);
    },

    toggleTheme: () => {
        const currentTheme = window.themeManager.normalizeTheme(document.documentElement.getAttribute("data-theme"));
        const nextTheme = currentTheme === "dark" ? "light" : "dark";
        window.themeManager.setTheme(nextTheme);
    },

    syncToggleUi: (activeTheme) => {
        const normalizedTheme = window.themeManager.normalizeTheme(activeTheme);
        const iconElements = document.querySelectorAll("[data-theme-toggle-icon]");
        const textElements = document.querySelectorAll("[data-theme-toggle-text]");

        const nextAction = normalizedTheme === "dark"
            ? { icon: "bi-sun", text: "Light mode" }
            : { icon: "bi-moon", text: "Dark mode" };

        for (const iconElement of iconElements) {
            iconElement.classList.remove("bi-sun", "bi-moon");
            iconElement.classList.add(nextAction.icon);
        }

        for (const textElement of textElements) {
            textElement.textContent = nextAction.text;
        }
    }
};

window.themeManager.applyTheme(window.themeManager.getSavedTheme());

document.addEventListener("click", (event) => {
    const toggleButton = event.target.closest("[data-theme-toggle]");

    if (!toggleButton) {
        return;
    }

    event.preventDefault();
    window.themeManager.toggleTheme();
});

document.addEventListener("DOMContentLoaded", () => {
    const activeTheme = document.documentElement.getAttribute("data-theme") === "dark" ? "dark" : "light";
    window.themeManager.syncToggleUi(activeTheme);
});
