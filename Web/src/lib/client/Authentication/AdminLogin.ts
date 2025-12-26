import { createEndpoint } from "../api";
export type AdminLoginRequest = {
    readonly username: string;
    readonly passwordHash: string;
};
export type AdminLoginResponse = {
    readonly token: string;
};
export const AdminLogin = createEndpoint<AdminLoginRequest, AdminLoginResponse>('/api/admin/login', 'POST');
