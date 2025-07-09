import { fileURLToPath, URL } from 'node:url';
import { defineConfig } from 'vite';
import plugin from '@vitejs/plugin-react';
import fs from 'fs';
import { env } from 'process';
import tailwindcss from '@tailwindcss/vite'

const baseFolder =
    env.APPDATA !== undefined && env.APPDATA !== ''
        ? `${env.APPDATA}/ASP.NET/https`
        : `${env.HOME}/.aspnet/https`;

if (!fs.existsSync(baseFolder)) {
    fs.mkdirSync(baseFolder, { recursive: true });
}

//const target = env.ASPNETCORE_HTTP_PORTS ? `https://localhost:${env.ASPNETCORE_HTTP_PORTS}` :
//    env.ASPNETCORE_URLS ? env.ASPNETCORE_URLS.split(';')[0] : 'http://localhost:32777';
////const target = 'http://localhost:32678';

const target = process.env.API_URL || 'http://localhost:32678';

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [
        plugin(),
        tailwindcss(),
    ],
    resolve: {
        alias: {
            '@': fileURLToPath(new URL('./src', import.meta.url))
        }
    },
    server: {
        // The following is only valid in dev environments, so we need to find another way for production.
        //proxy: {
        //    '^/weatherforecast': {
        //        target,
        //        secure: false
        //    },
        //    '^/user': {
        //        target,
        //        secure: false,
        //        changeOrigin: true
        //    },            
        //},
        port: parseInt(env.DEV_SERVER_PORT || '51144'),
        host: '0.0.0.0',
        watch: {
            usePolling: true,
            interval: 1000, //in ms
        }
    }
})
