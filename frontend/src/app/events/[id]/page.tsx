"use client";

import { useQuery } from "@tanstack/react-query";
import { useParams, useRouter } from "next/navigation";
import { useState } from "react";
import { Calendar, MapPin, Clock, User, Minus, Plus } from "lucide-react";
import { getEventById } from "@/lib/api/events";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import type { TicketTier } from "@/types/api";

interface TicketSelection {
  ticket_tier_id: string;
  quantity: number;
  price: number;
}

export default function EventDetailPage() {
  const params = useParams();
  const router = useRouter();
  const eventId = params.id as string;

  const [selections, setSelections] = useState<Record<string, number>>({});

  const { data: event, isLoading, error } = useQuery({
    queryKey: ["event", eventId],
    queryFn: () => getEventById(eventId),
  });

  const handleQuantityChange = (tierId: string, change: number) => {
    setSelections((prev) => {
      const current = prev[tierId] || 0;
      const newValue = Math.max(0, current + change);

      if (newValue === 0) {
        const { [tierId]: _, ...rest } = prev;
        return rest;
      }

      return { ...prev, [tierId]: newValue };
    });
  };

  const getTotalAmount = () => {
    if (!event) return 0;

    return Object.entries(selections).reduce((total, [tierId, quantity]) => {
      const tier = event.ticket_tiers.find((t) => t.id === tierId);
      return total + (tier?.price || 0) * quantity;
    }, 0);
  };

  const getTotalTickets = () => {
    return Object.values(selections).reduce((sum, qty) => sum + qty, 0);
  };

  const handleCheckout = () => {
    const items = Object.entries(selections).map(([ticket_tier_id, quantity]) => ({
      ticket_tier_id,
      quantity,
    }));

    // Store order data in sessionStorage for checkout page
    sessionStorage.setItem("checkout_data", JSON.stringify({
      event_id: eventId,
      items,
      event_title: event?.title,
      event_date: event?.start_date,
      event_location: event?.location,
    }));

    router.push("/checkout");
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

  const formatTime = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleTimeString("id-ID", {
      hour: "2-digit",
      minute: "2-digit",
      timeZone: event?.timezone || "Asia/Jakarta",
    });
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
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Memuat detail event...</p>
        </div>
      </div>
    );
  }

  if (error || !event) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 mb-4">Gagal memuat detail event</p>
          <Button onClick={() => router.push("/")}>Kembali ke Beranda</Button>
        </div>
      </div>
    );
  }

  const totalTickets = getTotalTickets();
  const totalAmount = getTotalAmount();
  const platformFee = Math.floor(totalAmount * 0.05);
  const serviceFee = totalTickets * 2500;
  const grandTotal = totalAmount + platformFee + serviceFee;

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Hero Section with Banner */}
      <div className="relative h-[400px] bg-gradient-to-r from-blue-600 to-purple-600">
        {event.banner_url ? (
          <img
            src={event.banner_url}
            alt={event.title}
            className="w-full h-full object-cover"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center">
            <Calendar className="h-32 w-32 text-white opacity-50" />
          </div>
        )}
        <div className="absolute inset-0 bg-black bg-opacity-40" />
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 -mt-32 relative z-10">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Content */}
          <div className="lg:col-span-2 space-y-6">
            {/* Event Info Card */}
            <Card className="p-8">
              <h1 className="text-4xl font-bold text-gray-900 mb-4">
                {event.title}
              </h1>

              <div className="space-y-4 text-gray-700">
                <div className="flex items-start gap-3">
                  <Calendar className="h-5 w-5 mt-1 text-blue-600" />
                  <div>
                    <p className="font-semibold">Tanggal & Waktu</p>
                    <p>{formatDate(event.start_date)}</p>
                    <p className="text-sm text-gray-600">
                      {formatTime(event.start_date)} - {formatTime(event.end_date)} {event.timezone}
                    </p>
                  </div>
                </div>

                <div className="flex items-start gap-3">
                  <MapPin className="h-5 w-5 mt-1 text-blue-600" />
                  <div>
                    <p className="font-semibold">Lokasi</p>
                    <p>{event.venue}</p>
                    <p className="text-sm text-gray-600">{event.location}</p>
                  </div>
                </div>

                {event.organizer_name && (
                  <div className="flex items-start gap-3">
                    <User className="h-5 w-5 mt-1 text-blue-600" />
                    <div>
                      <p className="font-semibold">Penyelenggara</p>
                      <p>{event.organizer_name}</p>
                    </div>
                  </div>
                )}
              </div>

              <div className="mt-8 pt-8 border-t">
                <h2 className="text-2xl font-bold mb-4">Tentang Event</h2>
                <div className="prose max-w-none">
                  <p className="text-gray-700 whitespace-pre-wrap">{event.description}</p>
                </div>
              </div>
            </Card>

            {/* Ticket Tiers */}
            <Card className="p-8">
              <h2 className="text-2xl font-bold mb-6">Pilih Tiket</h2>

              {event.ticket_tiers && event.ticket_tiers.length > 0 ? (
                <div className="space-y-4">
                  {event.ticket_tiers.map((tier) => {
                    const isAvailable = tier.available_count > 0;
                    const selectedQty = selections[tier.id] || 0;
                    const maxQty = Math.min(tier.available_count, 10);

                    return (
                      <div
                        key={tier.id}
                        className={`border rounded-lg p-6 ${
                          !isAvailable ? "bg-gray-50 opacity-60" : "bg-white"
                        }`}
                      >
                        <div className="flex items-start justify-between mb-4">
                          <div className="flex-1">
                            <h3 className="text-xl font-semibold text-gray-900">
                              {tier.name}
                            </h3>
                            {tier.description && (
                              <p className="text-sm text-gray-600 mt-1">
                                {tier.description}
                              </p>
                            )}
                            <p className="text-2xl font-bold text-blue-600 mt-2">
                              {formatCurrency(tier.price)}
                            </p>
                          </div>
                        </div>

                        <div className="flex items-center justify-between">
                          <div className="text-sm">
                            {isAvailable ? (
                              <span className="text-green-600 font-medium">
                                {tier.available_count} tiket tersedia
                              </span>
                            ) : (
                              <span className="text-red-600 font-medium">
                                Tiket habis
                              </span>
                            )}
                          </div>

                          {isAvailable && (
                            <div className="flex items-center gap-3">
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => handleQuantityChange(tier.id, -1)}
                                disabled={selectedQty === 0}
                              >
                                <Minus className="h-4 w-4" />
                              </Button>
                              <span className="w-12 text-center font-semibold">
                                {selectedQty}
                              </span>
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => handleQuantityChange(tier.id, 1)}
                                disabled={selectedQty >= maxQty}
                              >
                                <Plus className="h-4 w-4" />
                              </Button>
                            </div>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              ) : (
                <p className="text-center text-gray-600 py-8">
                  Belum ada tiket yang tersedia untuk event ini
                </p>
              )}
            </Card>
          </div>

          {/* Sidebar - Order Summary */}
          <div className="lg:col-span-1">
            <Card className="p-6 sticky top-4">
              <h3 className="text-xl font-bold mb-4">Ringkasan Pesanan</h3>

              {totalTickets === 0 ? (
                <p className="text-center text-gray-500 py-8">
                  Pilih tiket untuk melanjutkan
                </p>
              ) : (
                <>
                  <div className="space-y-3 mb-6">
                    {Object.entries(selections).map(([tierId, quantity]) => {
                      const tier = event.ticket_tiers.find((t) => t.id === tierId);
                      if (!tier) return null;

                      return (
                        <div key={tierId} className="flex justify-between text-sm">
                          <span className="text-gray-700">
                            {tier.name} x{quantity}
                          </span>
                          <span className="font-semibold">
                            {formatCurrency(tier.price * quantity)}
                          </span>
                        </div>
                      );
                    })}
                  </div>

                  <div className="space-y-2 py-4 border-t border-gray-200 text-sm">
                    <div className="flex justify-between">
                      <span className="text-gray-600">Subtotal</span>
                      <span className="font-medium">{formatCurrency(totalAmount)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Platform Fee (5%)</span>
                      <span className="font-medium">{formatCurrency(platformFee)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Service Fee</span>
                      <span className="font-medium">{formatCurrency(serviceFee)}</span>
                    </div>
                  </div>

                  <div className="pt-4 border-t border-gray-200 mb-6">
                    <div className="flex justify-between items-center">
                      <span className="text-lg font-bold">Total</span>
                      <span className="text-2xl font-bold text-blue-600">
                        {formatCurrency(grandTotal)}
                      </span>
                    </div>
                  </div>

                  <Button
                    className="w-full"
                    size="lg"
                    onClick={handleCheckout}
                  >
                    Lanjutkan ke Pembayaran
                  </Button>

                  <p className="text-xs text-gray-500 text-center mt-4">
                    Tiket akan direservasi selama 15 menit setelah checkout
                  </p>
                </>
              )}
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}
