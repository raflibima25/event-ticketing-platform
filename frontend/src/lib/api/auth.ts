import { post, get } from "./client";
import type {
  LoginRequest,
  RegisterRequest,
  AuthResponse,
  User,
  ChangePasswordRequest,
  ForgotPasswordRequest,
  ResetPasswordRequest,
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

/**
 * Change password for authenticated user
 */
export async function changePassword(data: ChangePasswordRequest): Promise<void> {
  await post<void, ChangePasswordRequest>("/auth/change-password", data);
}

/**
 * Request password reset email
 */
export async function forgotPassword(data: ForgotPasswordRequest): Promise<void> {
  await post<void, ForgotPasswordRequest>("/auth/forgot-password", data);
}

/**
 * Reset password with token
 */
export async function resetPassword(data: ResetPasswordRequest): Promise<void> {
  await post<void, ResetPasswordRequest>("/auth/reset-password", data);
}
