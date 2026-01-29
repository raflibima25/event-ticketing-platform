"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { resetPassword } from "@/lib/api/auth";

const resetPasswordSchema = z
  .object({
    new_password: z.string().min(8, "Password minimal 8 karakter"),
    confirm_password: z.string().min(8, "Konfirmasi password minimal 8 karakter"),
  })
  .refine((data) => data.new_password === data.confirm_password, {
    message: "Password tidak cocok",
    path: ["confirm_password"],
  });

type ResetPasswordFormData = z.infer<typeof resetPasswordSchema>;

export function ResetPasswordForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get("token");

  const [error, setError] = useState<string>("");
  const [success, setSuccess] = useState<boolean>(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ResetPasswordFormData>({
    resolver: zodResolver(resetPasswordSchema),
  });

  // If no token, show error
  if (!token) {
    return (
      <div className="space-y-4">
        <div className="p-4 text-sm text-red-600 bg-red-50 border border-red-200 rounded-lg">
          <p className="font-medium">Link Tidak Valid</p>
          <p className="mt-1">
            Link reset password tidak valid atau sudah kadaluarsa. Silakan
            request link baru.
          </p>
        </div>

        <p className="text-center text-sm text-gray-600">
          <Link href="/forgot-password" className="text-blue-600 hover:underline">
            Request link reset password baru
          </Link>
        </p>
      </div>
    );
  }

  const onSubmit = async (data: ResetPasswordFormData) => {
    try {
      setError("");
      await resetPassword({
        token,
        new_password: data.new_password,
      });
      setSuccess(true);

      // Redirect to login after 3 seconds
      setTimeout(() => {
        router.push("/login");
      }, 3000);
    } catch (err: unknown) {
      const errorMessage =
        err && typeof err === "object" && "message" in err
          ? String(err.message)
          : "Terjadi kesalahan saat reset password";
      setError(errorMessage);
    }
  };

  if (success) {
    return (
      <div className="space-y-4">
        <div className="p-4 text-sm text-green-600 bg-green-50 border border-green-200 rounded-lg">
          <p className="font-medium">Password Berhasil Direset!</p>
          <p className="mt-1">
            Password Anda telah berhasil diubah. Anda akan dialihkan ke halaman
            login...
          </p>
        </div>

        <p className="text-center text-sm text-gray-600">
          <Link href="/login" className="text-blue-600 hover:underline">
            Klik di sini jika tidak dialihkan otomatis
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
        label="Password Baru"
        type="password"
        placeholder="Masukkan password baru"
        error={errors.new_password?.message}
        {...register("new_password")}
      />

      <Input
        label="Konfirmasi Password"
        type="password"
        placeholder="Masukkan ulang password baru"
        error={errors.confirm_password?.message}
        {...register("confirm_password")}
      />

      <Button type="submit" className="w-full" isLoading={isSubmitting}>
        Reset Password
      </Button>

      <p className="text-center text-sm text-gray-600">
        <Link href="/login" className="text-blue-600 hover:underline">
          Kembali ke login
        </Link>
      </p>
    </form>
  );
}
