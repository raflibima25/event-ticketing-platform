"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { register as registerUser } from "@/lib/api/auth";
import { useAuthStore } from "@/store/authStore";
import type { RegisterRequest } from "@/types/api";

const registerSchema = z
  .object({
    email: z.string().email("Email tidak valid"),
    password: z.string().min(6, "Password minimal 6 karakter"),
    confirm_password: z.string(),
    full_name: z.string().min(3, "Nama lengkap minimal 3 karakter"),
    phone: z
      .string()
      .regex(/^(\+62|62|0)[0-9]{9,12}$/, "Nomor telepon tidak valid"),
    role: z.enum(["customer", "organizer"]).default("customer"),
  })
  .refine((data) => data.password === data.confirm_password, {
    message: "Password tidak sama",
    path: ["confirm_password"],
  });

type RegisterFormData = z.infer<typeof registerSchema>;

export function RegisterForm() {
  const router = useRouter();
  const { setUser, setTokens } = useAuthStore();
  const [error, setError] = useState<string>("");

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    watch,
  } = useForm<RegisterFormData>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      role: "customer",
    },
  });

  const selectedRole = watch("role");

  const onSubmit = async (data: RegisterFormData) => {
    try {
      setError("");
      const { confirm_password, ...registerData } = data;
      const response = await registerUser(registerData as RegisterRequest);

      // Save tokens and user to store
      setTokens(response.access_token, response.refresh_token);
      setUser(response.user);

      // Save user to localStorage for persistence
      if (typeof window !== "undefined") {
        localStorage.setItem("user", JSON.stringify(response.user));
      }

      // Redirect based on role
      if (response.user.role === "organizer") {
        router.push("/organizer/dashboard");
      } else {
        router.push("/");
      }
    } catch (err: unknown) {
      const errorMessage =
        err && typeof err === "object" && "message" in err
          ? String(err.message)
          : "Terjadi kesalahan saat registrasi";
      setError(errorMessage);
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      {error && (
        <div className="p-3 text-sm text-red-600 bg-red-50 border border-red-200 rounded-lg">
          {error}
        </div>
      )}

      <Input
        label="Nama Lengkap"
        type="text"
        placeholder="John Doe"
        error={errors.full_name?.message}
        {...register("full_name")}
      />

      <Input
        label="Email"
        type="email"
        placeholder="nama@email.com"
        error={errors.email?.message}
        {...register("email")}
      />

      <Input
        label="Nomor Telepon"
        type="tel"
        placeholder="08123456789"
        error={errors.phone?.message}
        {...register("phone")}
      />

      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Daftar Sebagai
        </label>
        <div className="flex gap-4">
          <label className="flex items-center">
            <input
              type="radio"
              value="customer"
              {...register("role")}
              className="mr-2"
            />
            <span className="text-sm">Customer</span>
          </label>
          <label className="flex items-center">
            <input
              type="radio"
              value="organizer"
              {...register("role")}
              className="mr-2"
            />
            <span className="text-sm">Event Organizer</span>
          </label>
        </div>
        {errors.role && (
          <p className="mt-1 text-sm text-red-600">{errors.role.message}</p>
        )}
      </div>

      <Input
        label="Password"
        type="password"
        placeholder="Minimal 6 karakter"
        error={errors.password?.message}
        {...register("password")}
      />

      <Input
        label="Konfirmasi Password"
        type="password"
        placeholder="Masukkan password lagi"
        error={errors.confirm_password?.message}
        {...register("confirm_password")}
      />

      <Button type="submit" className="w-full" isLoading={isSubmitting}>
        Daftar
      </Button>

      <p className="text-center text-sm text-gray-600">
        Sudah punya akun?{" "}
        <Link href="/login" className="text-blue-600 hover:underline">
          Masuk di sini
        </Link>
      </p>
    </form>
  );
}
