"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { Clock, CheckCircle, XCircle, AlertCircle, Loader2, ExternalLink } from "lucide-react";
import { getOrderById } from "@/lib/api/orders";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";

export default function PaymentPage() {
  const params = useParams();
  const router = useRouter();
  const orderId = params.orderId as string;

  const [timeLeft, setTimeLeft] = useState<number>(0);

  const { data: order, isLoading, error, refetch } = useQuery({
    queryKey: ["order", orderId],
    queryFn: () => getOrderById(orderId),
    refetchInterval: (data) => {
      // Poll every 5 seconds if order is still in reserved status
      return data?.status === "reserved" ? 5000 : false;
    },
  });

  useEffect(() => {
    if (!order) return;

    // Calculate time left for reservation
    const expiryDate = new Date(order.reservation_expires_at);
    const now = new Date();
    const diff = expiryDate.getTime() - now.getTime();

    if (diff > 0) {
      setTimeLeft(Math.floor(diff / 1000));

      const interval = setInterval(() => {
        const newDiff = expiryDate.getTime() - new Date().getTime();
        if (newDiff <= 0) {
          setTimeLeft(0);
          clearInterval(interval);
          refetch();
        } else {
          setTimeLeft(Math.floor(newDiff / 1000));
        }
      }, 1000);

      return () => clearInterval(interval);
    }
  }, [order, refetch]);

  const formatTime = (seconds: number) => {
    const minutes = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${minutes}:${secs.toString().padStart(2, "0")}`;
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(amount);
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="h-12 w-12 animate-spin text-blue-600 mx-auto mb-4" />
          <p className="text-gray-600">Memuat data pembayaran...</p>
        </div>
      </div>
    );
  }

  if (error || !order) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <XCircle className="h-16 w-16 text-red-600 mx-auto mb-4" />
          <p className="text-xl font-semibold text-gray-900 mb-2">Pesanan Tidak Ditemukan</p>
          <p className="text-gray-600 mb-6">Pesanan yang Anda cari tidak ditemukan</p>
          <Button onClick={() => router.push("/")}>Kembali ke Beranda</Button>
        </div>
      </div>
    );
  }

  // Payment success
  if (order.status === "paid") {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
        <Card className="max-w-md w-full p-8 text-center">
          <CheckCircle className="h-16 w-16 text-green-600 mx-auto mb-4" />
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Pembayaran Berhasil!</h1>
          <p className="text-gray-600 mb-6">
            Terima kasih atas pembayaran Anda. E-Ticket telah dikirim ke email Anda.
          </p>
          <div className="space-y-3">
            <Button className="w-full" onClick={() => router.push("/tickets")}>
              Lihat Tiket Saya
            </Button>
            <Button
              variant="outline"
              className="w-full"
              onClick={() => router.push("/orders")}
            >
              Lihat Pesanan
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  // Payment expired or cancelled
  if (order.status === "expired" || order.status === "cancelled") {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 p-4">
        <Card className="max-w-md w-full p-8 text-center">
          <XCircle className="h-16 w-16 text-red-600 mx-auto mb-4" />
          <h1 className="text-2xl font-bold text-gray-900 mb-2">
            {order.status === "expired" ? "Pesanan Kadaluarsa" : "Pesanan Dibatalkan"}
          </h1>
          <p className="text-gray-600 mb-6">
            {order.status === "expired"
              ? "Waktu pembayaran telah habis. Tiket telah dikembalikan ke kuota."
              : "Pesanan Anda telah dibatalkan."}
          </p>
          <Button onClick={() => router.push("/")}>Kembali ke Beranda</Button>
        </Card>
      </div>
    );
  }

  // Waiting for payment
  return (
    <div className="min-h-screen bg-gray-50 py-12">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Menunggu Pembayaran</h1>
          <p className="text-gray-600 mt-2">
            Selesaikan pembayaran Anda sebelum waktu habis
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Payment Instructions */}
          <div className="lg:col-span-2 space-y-6">
            {/* Timer Card */}
            <Card className="p-6 bg-gradient-to-r from-blue-600 to-purple-600 text-white">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm opacity-90">Waktu Tersisa</p>
                  <p className="text-3xl font-bold mt-1">{formatTime(timeLeft)}</p>
                </div>
                <Clock className="h-12 w-12 opacity-80" />
              </div>
              {timeLeft < 300 && (
                <div className="mt-4 bg-white/20 rounded-lg p-3 flex items-start gap-2">
                  <AlertCircle className="h-5 w-5 flex-shrink-0 mt-0.5" />
                  <p className="text-sm">
                    Segera selesaikan pembayaran! Pesanan akan dibatalkan secara otomatis jika waktu habis.
                  </p>
                </div>
              )}
            </Card>

            {/* Payment Details Card */}
            <Card className="p-8">
              <h2 className="text-xl font-bold mb-6">Detail Pembayaran</h2>

              <div className="space-y-4">
                <div className="flex justify-between pb-4 border-b">
                  <span className="text-gray-600">Order ID</span>
                  <span className="font-mono font-semibold">{order.id.substring(0, 8)}</span>
                </div>

                {order.items.map((item, index) => (
                  <div key={index} className="flex justify-between text-sm">
                    <span className="text-gray-700">
                      Tiket x{item.quantity}
                    </span>
                    <span className="font-semibold">{formatCurrency(item.subtotal)}</span>
                  </div>
                ))}

                <div className="pt-4 border-t space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-gray-600">Subtotal</span>
                    <span className="font-medium">{formatCurrency(order.total_amount)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">Platform Fee</span>
                    <span className="font-medium">{formatCurrency(order.platform_fee)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">Service Fee</span>
                    <span className="font-medium">{formatCurrency(order.service_fee)}</span>
                  </div>
                </div>

                <div className="pt-4 border-t">
                  <div className="flex justify-between items-center">
                    <span className="text-lg font-bold">Total</span>
                    <span className="text-2xl font-bold text-blue-600">
                      {formatCurrency(order.grand_total)}
                    </span>
                  </div>
                </div>
              </div>
            </Card>

            {/* Payment Instructions */}
            <Card className="p-8">
              <h2 className="text-xl font-bold mb-4">Cara Pembayaran</h2>

              <div className="space-y-4 text-gray-700">
                <div>
                  <p className="font-semibold mb-2">1. Klik tombol "Bayar Sekarang"</p>
                  <p className="text-sm">
                    Anda akan diarahkan ke halaman pembayaran Xendit untuk memilih metode pembayaran.
                  </p>
                </div>

                <div>
                  <p className="font-semibold mb-2">2. Pilih metode pembayaran</p>
                  <p className="text-sm">
                    Pilih metode yang Anda inginkan: QRIS, Virtual Account, atau E-Wallet.
                  </p>
                </div>

                <div>
                  <p className="font-semibold mb-2">3. Selesaikan pembayaran</p>
                  <p className="text-sm">
                    Ikuti instruksi pembayaran sesuai metode yang dipilih. Pastikan membayar sebelum waktu habis.
                  </p>
                </div>

                <div>
                  <p className="font-semibold mb-2">4. Dapatkan e-ticket</p>
                  <p className="text-sm">
                    Setelah pembayaran berhasil, e-ticket akan dikirim ke email Anda dan dapat dilihat di halaman "Tiket Saya".
                  </p>
                </div>
              </div>
            </Card>
          </div>

          {/* Actions Sidebar */}
          <div className="lg:col-span-1">
            <Card className="p-6 sticky top-4">
              <div className="space-y-4">
                {/* Payment button would go here - for now showing placeholder */}
                <Button className="w-full" size="lg" disabled>
                  <ExternalLink className="h-5 w-5 mr-2" />
                  Bayar Sekarang
                </Button>

                <p className="text-xs text-center text-gray-500">
                  Halaman pembayaran akan tersedia setelah integrasi Xendit lengkap
                </p>

                <div className="pt-4 border-t">
                  <Button
                    variant="outline"
                    className="w-full"
                    onClick={() => refetch()}
                  >
                    Cek Status Pembayaran
                  </Button>
                </div>

                <div className="pt-4 border-t">
                  <Button
                    variant="ghost"
                    className="w-full text-red-600 hover:text-red-700 hover:bg-red-50"
                    onClick={() => {
                      if (confirm("Apakah Anda yakin ingin membatalkan pesanan?")) {
                        router.push("/");
                      }
                    }}
                  >
                    Batalkan Pesanan
                  </Button>
                </div>
              </div>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}
