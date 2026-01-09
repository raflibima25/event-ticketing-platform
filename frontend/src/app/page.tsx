"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Navbar } from "@/components/layout/navbar";
import { EventCard } from "@/components/events/event-card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { getEvents } from "@/lib/api/events";
import { Search, Loader2 } from "lucide-react";

export default function Home() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [searchInput, setSearchInput] = useState("");

  const { data, isLoading, error } = useQuery({
    queryKey: ["events", page, search],
    queryFn: () => getEvents({ page, limit: 12, search, status: "published" }),
  });

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setSearch(searchInput);
    setPage(1);
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <Navbar />

      <main>
        {/* Hero Section */}
        <div className="bg-gradient-to-r from-blue-600 to-purple-600 text-white">
          <div className="container mx-auto px-4 py-16">
            <div className="max-w-3xl mx-auto text-center">
              <h1 className="text-4xl md:text-5xl font-bold mb-4">
                Temukan Event Menarik
              </h1>
              <p className="text-lg md:text-xl mb-8 opacity-90">
                Platform tiket event terpercaya untuk berbagai acara menarik di
                Indonesia
              </p>

              {/* Search Bar */}
              <form onSubmit={handleSearch} className="max-w-2xl mx-auto">
                <div className="flex gap-2">
                  <div className="flex-1">
                    <Input
                      type="text"
                      placeholder="Cari event berdasarkan nama..."
                      value={searchInput}
                      onChange={(e) => setSearchInput(e.target.value)}
                      className="bg-white"
                    />
                  </div>
                  <Button type="submit" size="lg">
                    <Search className="h-5 w-5 mr-2" />
                    Cari
                  </Button>
                </div>
              </form>
            </div>
          </div>
        </div>

        {/* Events List */}
        <div className="container mx-auto px-4 py-12">
          <div className="mb-8">
            <h2 className="text-2xl font-bold text-gray-900">
              {search ? `Hasil Pencarian: "${search}"` : "Semua Event"}
            </h2>
          </div>

          {isLoading && (
            <div className="flex items-center justify-center py-20">
              <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
              <span className="ml-2 text-gray-600">Memuat event...</span>
            </div>
          )}

          {error && (
            <div className="text-center py-20">
              <p className="text-red-600">
                Gagal memuat event. Silakan coba lagi.
              </p>
            </div>
          )}

          {data && data.events && data.events.length === 0 && (
            <div className="text-center py-20">
              <p className="text-gray-600">Tidak ada event yang ditemukan.</p>
            </div>
          )}

          {data && data.events && data.events.length > 0 && (
            <>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {data.events.map((event) => (
                  <EventCard key={event.id} event={event} />
                ))}
              </div>

              {/* Pagination */}
              {data.meta && data.meta.total_pages > 1 && (
                <div className="mt-12 flex items-center justify-center gap-2">
                  <Button
                    variant="outline"
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                    disabled={page === 1}
                  >
                    Sebelumnya
                  </Button>

                  <div className="flex items-center gap-2">
                    {Array.from(
                      { length: data?.meta?.total_pages || 1 },
                      (_, i) => i + 1
                    )
                      .filter(
                        (p) =>
                          p === 1 ||
                          p === (data?.meta?.total_pages || 1) ||
                          Math.abs(p - page) <= 1
                      )
                      .map((p, idx, arr) => {
                        const prevPage = arr[idx - 1];
                        const showEllipsis = prevPage && p - prevPage > 1;

                        return (
                          <div key={p} className="flex items-center gap-2">
                            {showEllipsis && (
                              <span className="text-gray-400">...</span>
                            )}
                            <Button
                              variant={p === page ? "primary" : "outline"}
                              onClick={() => setPage(p)}
                              size="sm"
                            >
                              {p}
                            </Button>
                          </div>
                        );
                      })}
                  </div>

                  <Button
                    variant="outline"
                    onClick={() =>
                      setPage((p) =>
                        Math.min(data?.meta?.total_pages || 1, p + 1)
                      )
                    }
                    disabled={page === data?.meta?.total_pages}
                  >
                    Selanjutnya
                  </Button>
                </div>
              )}
            </>
          )}
        </div>
      </main>

      {/* Footer */}
      <footer className="bg-gray-900 text-white py-8 mt-20">
        <div className="container mx-auto px-4 text-center">
          <p>&copy; 2025 Event Ticketing Platform. All rights reserved.</p>
        </div>
      </footer>
    </div>
  );
}
