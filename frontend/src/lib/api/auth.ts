import { post, get } from "./client";
import type {
  LoginRequest,
  RegisterRequest,
  AuthResponse,
  User,
} from "@/types/api";

/**
 * Login user
 */
export async function login(data: LoginRequest): Promise<AuthResponse> {
  return post<AuthResponse, LoginRequest>("/auth/login", data);
}

/**
 * Register new user
 */
export async function register(data: RegisterRequest): Promise<AuthResponse> {
  return post<AuthResponse, RegisterRequest>("/auth/register", data);
}

/**
 * Get current user profile
 */
export async function getProfile(): Promise<User> {
  return get<User>("/auth/profile");
}

/**
 * Logout user (client-side only - clears tokens)
 */
export function logout(): void {
  if (typeof window !== "undefined") {
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
    localStorage.removeItem("user");
    window.location.href = "/login";
  }
}
