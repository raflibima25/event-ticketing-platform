"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useMutation } from "@tanstack/react-query";
import { Calendar, MapPin, AlertCircle, Loader2 } from "lucide-react";
import { createOrder } from "@/lib/api/orders";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import type { CreateOrderRequest } from "@/types/api";

interface CheckoutData {
  event_id: string;
  items: { ticket_tier_id: string; quantity: number }[];
  event_title: string;
  event_date: string;
  event_location: string;
}

export default function CheckoutPage() {
  const router = useRouter();
  const [checkoutData, setCheckoutData] = useState<CheckoutData | null>(null);
  const [customerName, setCustomerName] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");

  useEffect(() => {
    // Load checkout data from sessionStorage
    const data = sessionStorage.getItem("checkout_data");
    if (!data) {
      router.push("/");
      return;
    }

    setCheckoutData(JSON.parse(data));

    // Pre-fill email from localStorage if logged in
    const user = localStorage.getItem("user");
    if (user) {
      const userData = JSON.parse(user);
      setEmail(userData.email || "");
      setCustomerName(userData.full_name || "");
      setPhone(userData.phone || "");
    }
  }, [router]);

  const createOrderMutation = useMutation({
    mutationFn: (data: CreateOrderRequest) => createOrder(data),
    onSuccess: (response) => {
      // Clear checkout data
      sessionStorage.removeItem("checkout_data");

      // Store order ID and navigate to payment page
      sessionStorage.setItem("order_id", response.order.id);
      router.push(`/payment/${response.order.id}`);
    },
    onError: (error: any) => {
      alert(error.message || "Gagal membuat pesanan. Silakan coba lagi.");
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!checkoutData) return;

    // Validate form
    if (!customerName || !email || !phone) {
      alert("Mohon lengkapi semua data");
      return;
    }

    // Create order
    createOrderMutation.mutate({
      event_id: checkoutData.event_id,
      items: checkoutData.items,
    });
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString("id-ID", {
      weekday: "long",
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  if (!checkoutData) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="h-12 w-12 animate-spin text-blue-600 mx-auto mb-4" />
          <p className="text-gray-600">Memuat data checkout...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-12">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Checkout</h1>
          <p className="text-gray-600 mt-2">
            Lengkapi data Anda untuk menyelesaikan pembelian tiket
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Form Section */}
          <div className="lg:col-span-2">
            <Card className="p-8">
              <h2 className="text-xl font-bold mb-6">Data Pembeli</h2>

              <form onSubmit={handleSubmit} className="space-y-6">
                <div>
                  <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-2">
                    Nama Lengkap <span className="text-red-500">*</span>
                  </label>
                  <Input
                    id="name"
                    type="text"
                    value={customerName}
                    onChange={(e) => setCustomerName(e.target.value)}
                    placeholder="Masukkan nama lengkap"
                    required
                  />
                </div>

                <div>
                  <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-2">
                    Email <span className="text-red-500">*</span>
                  </label>
                  <Input
                    id="email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="contoh@email.com"
                    required
                  />
                  <p className="text-xs text-gray-500 mt-1">
                    E-Ticket akan dikirim ke email ini
                  </p>
                </div>

                <div>
                  <label htmlFor="phone" className="block text-sm font-medium text-gray-700 mb-2">
                    No. Telepon <span className="text-red-500">*</span>
                  </label>
                  <Input
                    id="phone"
                    type="tel"
                    value={phone}
                    onChange={(e) => setPhone(e.target.value)}
                    placeholder="08123456789"
                    required
                  />
                </div>

                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 flex items-start gap-3">
                  <AlertCircle className="h-5 w-5 text-blue-600 mt-0.5 flex-shrink-0" />
                  <div className="text-sm text-blue-900">
                    <p className="font-semibold mb-1">Penting!</p>
                    <ul className="list-disc list-inside space-y-1">
                      <li>Tiket akan direservasi selama 15 menit</li>
                      <li>Setelah memilih metode pembayaran, waktu diperpanjang menjadi 30 menit</li>
                      <li>Lakukan pembayaran sebelum waktu habis</li>
                      <li>E-Ticket akan dikirim ke email setelah pembayaran berhasil</li>
                    </ul>
                  </div>
                </div>

                <Button
                  type="submit"
                  className="w-full"
                  size="lg"
                  disabled={createOrderMutation.isPending}
                >
                  {createOrderMutation.isPending ? (
                    <>
                      <Loader2 className="h-5 w-5 animate-spin mr-2" />
                      Memproses...
                    </>
                  ) : (
                    "Lanjut ke Pembayaran"
                  )}
                </Button>
              </form>
            </Card>
          </div>

          {/* Summary Section */}
          <div className="lg:col-span-1">
            <Card className="p-6 sticky top-4">
              <h3 className="text-lg font-bold mb-4">Detail Event</h3>

              <div className="space-y-4">
                <div>
                  <h4 className="font-semibold text-gray-900">
                    {checkoutData.event_title}
                  </h4>
                </div>

                <div className="flex items-start gap-2 text-sm text-gray-600">
                  <Calendar className="h-4 w-4 mt-0.5 flex-shrink-0" />
                  <span>{formatDate(checkoutData.event_date)}</span>
                </div>

                <div className="flex items-start gap-2 text-sm text-gray-600">
                  <MapPin className="h-4 w-4 mt-0.5 flex-shrink-0" />
                  <span>{checkoutData.event_location}</span>
                </div>

                <div className="pt-4 border-t">
                  <div className="flex justify-between text-sm mb-2">
                    <span className="text-gray-600">Jumlah Tiket</span>
                    <span className="font-semibold">
                      {checkoutData.items.reduce((sum, item) => sum + item.quantity, 0)} tiket
                    </span>
                  </div>
                </div>
              </div>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}
