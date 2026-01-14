"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { Clock, CheckCircle, XCircle, Calendar, Loader2, ChevronRight } from "lucide-react";
import { getUserOrders } from "@/lib/api/orders";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import type { Order } from "@/types/api";

export default function OrdersPage() {
  const router = useRouter();
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [page, setPage] = useState(1);
  const limit = 10;

  useEffect(() => {
    const token = localStorage.getItem("access_token");
    if (!token) {
      router.push("/login");
      return;
    }
    setIsLoggedIn(true);
  }, [router]);

  const { data, isLoading, error } = useQuery({
    queryKey: ["orders", page],
    queryFn: () => getUserOrders({ page, limit }),
    enabled: isLoggedIn,
  });

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(amount);
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString("id-ID", {
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const getStatusBadge = (status: Order["status"]) => {
    switch (status) {
      case "paid":
        return (
          <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full text-sm font-medium bg-green-100 text-green-800">
            <CheckCircle className="h-4 w-4" />
            Dibayar
          </span>
        );
      case "reserved":
        return (
          <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full text-sm font-medium bg-blue-100 text-blue-800">
            <Clock className="h-4 w-4" />
            Menunggu Pembayaran
          </span>
        );
      case "expired":
        return (
          <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full text-sm font-medium bg-gray-100 text-gray-800">
            <XCircle className="h-4 w-4" />
            Kadaluarsa
          </span>
        );
      case "cancelled":
        return (
          <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full text-sm font-medium bg-red-100 text-red-800">
            <XCircle className="h-4 w-4" />
            Dibatalkan
          </span>
        );
      default:
        return null;
    }
  };

  if (!isLoggedIn || isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="h-12 w-12 animate-spin text-blue-600 mx-auto mb-4" />
          <p className="text-gray-600">Memuat pesanan...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 mb-4">Gagal memuat pesanan</p>
          <Button onClick={() => router.push("/")}>Kembali ke Beranda</Button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-12">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Riwayat Pesanan</h1>
          <p className="text-gray-600 mt-2">
            Lihat semua transaksi dan status pembelian tiket Anda
          </p>
        </div>

        {data && data.orders && data.orders.length > 0 ? (
          <>
            <div className="space-y-4">
              {data.orders.map((order) => (
                <Card key={order.id} className="p-6 hover:shadow-lg transition-shadow">
                  <div className="flex items-start justify-between mb-4">
                    <div>
                      <div className="flex items-center gap-3 mb-2">
                        <p className="font-mono text-sm text-gray-600">
                          Order #{order.id.substring(0, 8)}
                        </p>
                        {getStatusBadge(order.status)}
                      </div>
                      <p className="text-sm text-gray-600 flex items-center gap-2">
                        <Calendar className="h-4 w-4" />
                        {formatDate(order.created_at)}
                      </p>
                    </div>
                    <div className="text-right">
                      <p className="text-sm text-gray-600 mb-1">Total Pembayaran</p>
                      <p className="text-2xl font-bold text-blue-600">
                        {formatCurrency(order.grand_total)}
                      </p>
                    </div>
                  </div>

                  <div className="border-t pt-4">
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
                      <div>
                        <p className="text-sm text-gray-500">Jumlah Tiket</p>
                        <p className="font-semibold">
                          {order.items.reduce((sum, item) => sum + item.quantity, 0)} tiket
                        </p>
                      </div>
                      {order.payment_method && (
                        <div>
                          <p className="text-sm text-gray-500">Metode Pembayaran</p>
                          <p className="font-semibold capitalize">{order.payment_method}</p>
                        </div>
                      )}
                      {order.status === "reserved" && (
                        <div>
                          <p className="text-sm text-gray-500">Batas Waktu Pembayaran</p>
                          <p className="font-semibold text-red-600">
                            {formatDate(order.reservation_expires_at)}
                          </p>
                        </div>
                      )}
                    </div>

                    <div className="flex gap-3">
                      <Button
                        variant="outline"
                        className="flex-1"
                        onClick={() => router.push(`/payment/${order.id}`)}
                      >
                        Lihat Detail
                        <ChevronRight className="h-4 w-4 ml-1" />
                      </Button>

                      {order.status === "reserved" && (
                        <Button
                          className="flex-1"
                          onClick={() => router.push(`/payment/${order.id}`)}
                        >
                          Lanjutkan Pembayaran
                        </Button>
                      )}

                      {order.status === "paid" && (
                        <Button
                          className="flex-1"
                          onClick={() => router.push("/tickets")}
                        >
                          Lihat Tiket
                        </Button>
                      )}
                    </div>
                  </div>
                </Card>
              ))}
            </div>

            {/* Pagination */}
            {data.meta && data.meta.total_pages > 1 && (
              <div className="mt-8 flex justify-center gap-2">
                <Button
                  variant="outline"
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page === 1}
                >
                  Previous
                </Button>
                <div className="flex items-center gap-2">
                  {Array.from({ length: data.meta.total_pages }, (_, i) => i + 1).map((p) => (
                    <Button
                      key={p}
                      variant={p === page ? "default" : "outline"}
                      onClick={() => setPage(p)}
                      className="w-10"
                    >
                      {p}
                    </Button>
                  ))}
                </div>
                <Button
                  variant="outline"
                  onClick={() => setPage((p) => Math.min(data.meta.total_pages, p + 1))}
                  disabled={page === data.meta.total_pages}
                >
                  Next
                </Button>
              </div>
            )}
          </>
        ) : (
          <Card className="p-12 text-center">
            <Calendar className="h-16 w-16 text-gray-400 mx-auto mb-4" />
            <h3 className="text-xl font-semibold text-gray-900 mb-2">
              Belum Ada Pesanan
            </h3>
            <p className="text-gray-600 mb-6">
              Anda belum melakukan pembelian tiket. Jelajahi event menarik dan beli tiket sekarang!
            </p>
            <Button onClick={() => router.push("/")}>
              Jelajahi Event
            </Button>
          </Card>
        )}
      </div>
    </div>
  );
}
