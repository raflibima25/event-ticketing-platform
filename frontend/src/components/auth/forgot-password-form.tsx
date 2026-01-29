"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { forgotPassword } from "@/lib/api/auth";

const forgotPasswordSchema = z.object({
  email: z.string().email("Email tidak valid"),
});

type ForgotPasswordFormData = z.infer<typeof forgotPasswordSchema>;

export function ForgotPasswordForm() {
  const [error, setError] = useState<string>("");
  const [success, setSuccess] = useState<boolean>(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ForgotPasswordFormData>({
    resolver: zodResolver(forgotPasswordSchema),
  });

  const onSubmit = async (data: ForgotPasswordFormData) => {
    try {
      setError("");
      setSuccess(false);
      await forgotPassword(data);
      setSuccess(true);
    } catch (err: unknown) {
      const errorMessage =
        err && typeof err === "object" && "message" in err
          ? String(err.message)
          : "Terjadi kesalahan saat mengirim email";
      setError(errorMessage);
    }
  };

  if (success) {
    return (
      <div className="space-y-4">
        <div className="p-4 text-sm text-green-600 bg-green-50 border border-green-200 rounded-lg">
          <p className="font-medium">Email Terkirim!</p>
          <p className="mt-1">
            Jika email terdaftar di sistem kami, Anda akan menerima link untuk
            reset password. Silakan cek inbox atau folder spam Anda.
          </p>
        </div>

        <p className="text-center text-sm text-gray-600">
          <Link href="/login" className="text-blue-600 hover:underline">
            Kembali ke halaman login
          </Link>
        </p>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      {error && (
        <div className="p-3 text-sm text-red-600 bg-red-50 border border-red-200 rounded-lg">
          {error}
        </div>
      )}

      <Input
        label="Email"
        type="email"
        placeholder="nama@email.com"
        error={errors.email?.message}
        {...register("email")}
      />

      <Button type="submit" className="w-full" isLoading={isSubmitting}>
        Kirim Link Reset Password
      </Button>

      <p className="text-center text-sm text-gray-600">
        Sudah ingat password?{" "}
        <Link href="/login" className="text-blue-600 hover:underline">
          Kembali ke login
        </Link>
      </p>
    </form>
  );
}
