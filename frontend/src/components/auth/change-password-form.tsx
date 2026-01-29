"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { changePassword } from "@/lib/api/auth";

const changePasswordSchema = z
  .object({
    current_password: z.string().min(1, "Password saat ini wajib diisi"),
    new_password: z.string().min(8, "Password baru minimal 8 karakter"),
    confirm_password: z.string().min(8, "Konfirmasi password minimal 8 karakter"),
  })
  .refine((data) => data.new_password === data.confirm_password, {
    message: "Password baru tidak cocok",
    path: ["confirm_password"],
  })
  .refine((data) => data.current_password !== data.new_password, {
    message: "Password baru harus berbeda dengan password saat ini",
    path: ["new_password"],
  });

type ChangePasswordFormData = z.infer<typeof changePasswordSchema>;

interface ChangePasswordFormProps {
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function ChangePasswordForm({ onSuccess, onCancel }: ChangePasswordFormProps) {
  const [error, setError] = useState<string>("");
  const [success, setSuccess] = useState<boolean>(false);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<ChangePasswordFormData>({
    resolver: zodResolver(changePasswordSchema),
  });

  const onSubmit = async (data: ChangePasswordFormData) => {
    try {
      setError("");
      setSuccess(false);
      await changePassword({
        current_password: data.current_password,
        new_password: data.new_password,
      });
      setSuccess(true);
      reset();

      // Call onSuccess callback after a short delay
      if (onSuccess) {
        setTimeout(onSuccess, 2000);
      }
    } catch (err: unknown) {
      const errorMessage =
        err && typeof err === "object" && "message" in err
          ? String(err.message)
          : "Terjadi kesalahan saat mengubah password";
      setError(errorMessage);
    }
  };

  if (success) {
    return (
      <div className="p-4 text-sm text-green-600 bg-green-50 border border-green-200 rounded-lg">
        <p className="font-medium">Password Berhasil Diubah!</p>
        <p className="mt-1">
          Password Anda telah berhasil diperbarui.
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
        label="Password Saat Ini"
        type="password"
        placeholder="Masukkan password saat ini"
        error={errors.current_password?.message}
        {...register("current_password")}
      />

      <Input
        label="Password Baru"
        type="password"
        placeholder="Masukkan password baru"
        error={errors.new_password?.message}
        {...register("new_password")}
      />

      <Input
        label="Konfirmasi Password Baru"
        type="password"
        placeholder="Masukkan ulang password baru"
        error={errors.confirm_password?.message}
        {...register("confirm_password")}
      />

      <div className="flex gap-3 pt-2">
        {onCancel && (
          <Button
            type="button"
            variant="outline"
            className="flex-1"
            onClick={onCancel}
          >
            Batal
          </Button>
        )}
        <Button type="submit" className="flex-1" isLoading={isSubmitting}>
          Simpan Password
        </Button>
      </div>
    </form>
  );
}
