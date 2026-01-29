import { Suspense } from "react";
import { ResetPasswordForm } from "@/components/auth/reset-password-form";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

function ResetPasswordFormWrapper() {
  return <ResetPasswordForm />;
}

export default function ResetPasswordPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900">
            Event Ticketing Platform
          </h1>
          <p className="mt-2 text-gray-600">
            Platform penjualan tiket event terpercaya
          </p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Reset Password</CardTitle>
            <CardDescription>
              Masukkan password baru untuk akun Anda
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Suspense
              fallback={
                <div className="flex justify-center py-8">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                </div>
              }
            >
              <ResetPasswordFormWrapper />
            </Suspense>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
