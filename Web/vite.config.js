import { fileURLToPath, URL } from 'node:url';
import { defineConfig } from 'vite';
import plugin from '@vitejs/plugin-react';
import fs from 'fs';
import { env } from 'process';

const baseFolder =
    env.APPDATA !== undefined && env.APPDATA !== ''
        ? `${env.APPDATA}/ASP.NET/https`
        : `${env.HOME}/.aspnet/https`;

if (!fs.existsSync(baseFolder)) {
    fs.mkdirSync(baseFolder, { recursive: true });
}

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [
        plugin()
    ],
    resolve: {
        alias: {
            '@': fileURLToPath(new URL('./src', import.meta.url))
        }
    },
    server: {
        port: parseInt(env.DEV_SERVER_PORT || '8081'),
        host: '0.0.0.0',
        watch: {
            usePolling: true,
            interval: 1000, //in ms
        },
        hmr: {
            clientPort: 80,
            host: 'localhost',
            overlay: false
        }
    }
});
