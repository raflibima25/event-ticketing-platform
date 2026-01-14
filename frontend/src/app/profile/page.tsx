"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { User, Mail, Phone, Calendar, Shield, Loader2 } from "lucide-react";
import { getProfile } from "@/lib/api/auth";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

export default function ProfilePage() {
  const router = useRouter();
  const [isLoggedIn, setIsLoggedIn] = useState(false);

  useEffect(() => {
    const token = localStorage.getItem("access_token");
    if (!token) {
      router.push("/login");
      return;
    }
    setIsLoggedIn(true);
  }, [router]);

  const { data: user, isLoading, error } = useQuery({
    queryKey: ["profile"],
    queryFn: getProfile,
    enabled: isLoggedIn,
  });

  const handleLogout = () => {
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
    localStorage.removeItem("user");
    router.push("/login");
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString("id-ID", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  if (!isLoggedIn || isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="h-12 w-12 animate-spin text-blue-600 mx-auto mb-4" />
          <p className="text-gray-600">Memuat profil...</p>
        </div>
      </div>
    );
  }

  if (error || !user) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 mb-4">Gagal memuat profil</p>
          <Button onClick={() => router.push("/")}>Kembali ke Beranda</Button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-12">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Profil Saya</h1>
          <p className="text-gray-600 mt-2">
            Kelola informasi akun dan preferensi Anda
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Profile Card */}
          <div className="lg:col-span-2">
            <Card className="p-8">
              <div className="flex items-center gap-6 mb-8 pb-8 border-b">
                <div className="h-20 w-20 bg-gradient-to-br from-blue-600 to-purple-600 rounded-full flex items-center justify-center">
                  <User className="h-10 w-10 text-white" />
                </div>
                <div>
                  <h2 className="text-2xl font-bold text-gray-900">
                    {user.full_name}
                  </h2>
                  <p className="text-gray-600 flex items-center gap-2 mt-1">
                    <Shield className="h-4 w-4" />
                    <span className="capitalize">{user.role}</span>
                  </p>
                </div>
              </div>

              <div className="space-y-6">
                <div>
                  <label className="block text-sm font-medium text-gray-500 mb-2">
                    Email
                  </label>
                  <div className="flex items-center gap-3 text-gray-900">
                    <Mail className="h-5 w-5 text-gray-400" />
                    <span>{user.email}</span>
                    {user.is_email_verified && (
                      <span className="text-xs bg-green-100 text-green-800 px-2 py-1 rounded">
                        Terverifikasi
                      </span>
                    )}
                  </div>
                </div>

                {user.phone && (
                  <div>
                    <label className="block text-sm font-medium text-gray-500 mb-2">
                      No. Telepon
                    </label>
                    <div className="flex items-center gap-3 text-gray-900">
                      <Phone className="h-5 w-5 text-gray-400" />
                      <span>{user.phone}</span>
                    </div>
                  </div>
                )}

                <div>
                  <label className="block text-sm font-medium text-gray-500 mb-2">
                    Member Sejak
                  </label>
                  <div className="flex items-center gap-3 text-gray-900">
                    <Calendar className="h-5 w-5 text-gray-400" />
                    <span>{formatDate(user.created_at)}</span>
                  </div>
                </div>
              </div>

              <div className="mt-8 pt-8 border-t flex gap-4">
                <Button variant="outline" className="flex-1" disabled>
                  Edit Profil
                </Button>
                <Button variant="outline" className="flex-1" disabled>
                  Ubah Password
                </Button>
              </div>
            </Card>
          </div>

          {/* Quick Actions */}
          <div className="lg:col-span-1">
            <Card className="p-6">
              <h3 className="text-lg font-bold mb-4">Menu</h3>
              <div className="space-y-3">
                <Button
                  variant="outline"
                  className="w-full justify-start"
                  onClick={() => router.push("/tickets")}
                >
                  <Calendar className="h-4 w-4 mr-2" />
                  Tiket Saya
                </Button>
                <Button
                  variant="outline"
                  className="w-full justify-start"
                  onClick={() => router.push("/orders")}
                >
                  <Mail className="h-4 w-4 mr-2" />
                  Riwayat Pesanan
                </Button>
                <Button
                  variant="outline"
                  className="w-full justify-start text-red-600 hover:text-red-700 hover:bg-red-50"
                  onClick={handleLogout}
                >
                  Logout
                </Button>
              </div>
            </Card>

            {user.role === "organizer" && (
              <Card className="p-6 mt-6">
                <h3 className="text-lg font-bold mb-4">Organizer</h3>
                <p className="text-sm text-gray-600 mb-4">
                  Kelola event dan lihat statistik penjualan tiket
                </p>
                <Button variant="outline" className="w-full" disabled>
                  Dashboard Organizer
                </Button>
              </Card>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
