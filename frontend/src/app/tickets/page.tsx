"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { QrCode, Calendar, MapPin, Ticket as TicketIcon, Loader2, Download, CheckCircle } from "lucide-react";
import { getUserTickets } from "@/lib/api/tickets";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import type { Ticket } from "@/types/api";

export default function TicketsPage() {
  const router = useRouter();
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [selectedTicket, setSelectedTicket] = useState<Ticket | null>(null);

  useEffect(() => {
    const token = localStorage.getItem("access_token");
    if (!token) {
      router.push("/login");
      return;
    }
    setIsLoggedIn(true);
  }, [router]);

  const { data: tickets, isLoading, error } = useQuery({
    queryKey: ["tickets"],
    queryFn: getUserTickets,
    enabled: isLoggedIn,
  });

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
    });
  };

  const getStatusBadge = (status: Ticket["status"]) => {
    switch (status) {
      case "active":
        return (
          <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full text-sm font-medium bg-green-100 text-green-800">
            <CheckCircle className="h-4 w-4" />
            Aktif
          </span>
        );
      case "used":
        return (
          <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full text-sm font-medium bg-gray-100 text-gray-800">
            <QrCode className="h-4 w-4" />
            Sudah Digunakan
          </span>
        );
      case "cancelled":
        return (
          <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full text-sm font-medium bg-red-100 text-red-800">
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
          <p className="text-gray-600">Memuat tiket...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 mb-4">Gagal memuat tiket</p>
          <Button onClick={() => router.push("/")}>Kembali ke Beranda</Button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-12">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Tiket Saya</h1>
          <p className="text-gray-600 mt-2">
            Kelola dan tampilkan tiket event Anda
          </p>
        </div>

        {tickets && tickets.length > 0 ? (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {tickets.map((ticket) => (
              <Card
                key={ticket.id}
                className={`overflow-hidden hover:shadow-xl transition-shadow cursor-pointer ${
                  selectedTicket?.id === ticket.id ? "ring-2 ring-blue-500" : ""
                }`}
                onClick={() => setSelectedTicket(ticket)}
              >
                {/* Ticket Header with Event Info */}
                <div className="bg-gradient-to-r from-blue-600 to-purple-600 p-6 text-white">
                  <div className="flex items-start justify-between mb-4">
                    <div className="flex-1">
                      <h3 className="text-xl font-bold mb-2">{ticket.event_title}</h3>
                      <div className="space-y-1 text-sm opacity-90">
                        <div className="flex items-center gap-2">
                          <Calendar className="h-4 w-4" />
                          <span>{formatDate(ticket.event_date)} • {formatTime(ticket.event_date)}</span>
                        </div>
                        <div className="flex items-center gap-2">
                          <MapPin className="h-4 w-4" />
                          <span>{ticket.event_location}</span>
                        </div>
                      </div>
                    </div>
                    <TicketIcon className="h-12 w-12 opacity-50" />
                  </div>
                </div>

                {/* Ticket Body */}
                <div className="p-6">
                  <div className="flex items-start justify-between mb-4">
                    <div>
                      <p className="text-sm text-gray-600 mb-1">Tipe Tiket</p>
                      <p className="font-bold text-lg">{ticket.ticket_tier_name}</p>
                    </div>
                    {getStatusBadge(ticket.status)}
                  </div>

                  <div className="border-t pt-4 space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-gray-600">Pemegang Tiket</span>
                      <span className="font-medium">{ticket.user_name}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Email</span>
                      <span className="font-medium">{ticket.user_email}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Ticket ID</span>
                      <span className="font-mono text-xs">{ticket.id.substring(0, 12)}</span>
                    </div>
                  </div>

                  {ticket.status === "active" && (
                    <Button
                      className="w-full mt-6"
                      onClick={(e) => {
                        e.stopPropagation();
                        setSelectedTicket(ticket);
                      }}
                    >
                      <QrCode className="h-4 w-4 mr-2" />
                      Tampilkan QR Code
                    </Button>
                  )}

                  {ticket.status === "used" && ticket.validated_at && (
                    <div className="mt-4 text-sm text-gray-600 text-center">
                      Digunakan pada {formatDate(ticket.validated_at)} • {formatTime(ticket.validated_at)}
                    </div>
                  )}
                </div>
              </Card>
            ))}
          </div>
        ) : (
          <Card className="p-12 text-center">
            <TicketIcon className="h-16 w-16 text-gray-400 mx-auto mb-4" />
            <h3 className="text-xl font-semibold text-gray-900 mb-2">
              Belum Ada Tiket
            </h3>
            <p className="text-gray-600 mb-6">
              Anda belum memiliki tiket. Beli tiket event untuk melihatnya di sini!
            </p>
            <Button onClick={() => router.push("/")}>
              Jelajahi Event
            </Button>
          </Card>
        )}

        {/* QR Code Modal */}
        {selectedTicket && selectedTicket.status === "active" && (
          <div
            className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
            onClick={() => setSelectedTicket(null)}
          >
            <Card
              className="max-w-md w-full p-8"
              onClick={(e) => e.stopPropagation()}
            >
              <div className="text-center mb-6">
                <h2 className="text-2xl font-bold text-gray-900 mb-2">
                  {selectedTicket.event_title}
                </h2>
                <p className="text-gray-600">{selectedTicket.ticket_tier_name}</p>
              </div>

              {/* QR Code Display */}
              <div className="bg-white p-6 rounded-lg shadow-inner mb-6 flex items-center justify-center">
                {selectedTicket.qr_code ? (
                  <img
                    src={selectedTicket.qr_code}
                    alt="QR Code"
                    className="w-64 h-64"
                  />
                ) : (
                  <div className="w-64 h-64 bg-gray-100 flex items-center justify-center">
                    <QrCode className="h-24 w-24 text-gray-400" />
                  </div>
                )}
              </div>

              <div className="text-center space-y-3">
                <p className="text-sm text-gray-600">
                  Tunjukkan QR Code ini di pintu masuk event
                </p>
                <div className="flex gap-3">
                  <Button variant="outline" className="flex-1" disabled>
                    <Download className="h-4 w-4 mr-2" />
                    Download
                  </Button>
                  <Button
                    variant="outline"
                    className="flex-1"
                    onClick={() => setSelectedTicket(null)}
                  >
                    Tutup
                  </Button>
                </div>
              </div>
            </Card>
          </div>
        )}
      </div>
    </div>
  );
}
