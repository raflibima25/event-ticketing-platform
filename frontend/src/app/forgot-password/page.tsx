import { ForgotPasswordForm } from "@/components/auth/forgot-password-form";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default function ForgotPasswordPage() {
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
            <CardTitle>Lupa Password?</CardTitle>
            <CardDescription>
              Masukkan email Anda dan kami akan mengirimkan link untuk reset
              password
            </CardDescription>
          </CardHeader>
          <CardContent>
            <ForgotPasswordForm />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
