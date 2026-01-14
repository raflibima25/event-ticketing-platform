"use client";

import Link from "next/link";
import { useAuthStore } from "@/store/authStore";
import { Button } from "@/components/ui/button";
import { logout } from "@/lib/api/auth";
import { useEffect, useState } from "react";
import { User, LogOut, Calendar, Ticket, ShoppingBag, ChevronDown } from "lucide-react";

export function Navbar() {
  const { user, isAuthenticated, initAuth } = useAuthStore();
  const [showUserMenu, setShowUserMenu] = useState(false);

  useEffect(() => {
    initAuth();
  }, [initAuth]);

  const handleLogout = () => {
    logout();
    setShowUserMenu(false);
  };

  return (
    <nav className="bg-white shadow-sm border-b">
      <div className="container mx-auto px-4">
        <div className="flex items-center justify-between h-16">
          <Link href="/" className="flex items-center gap-2">
            <Calendar className="h-6 w-6 text-blue-600" />
            <span className="text-xl font-bold text-gray-900">
              Event Ticketing
            </span>
          </Link>

          <div className="flex items-center gap-4">
            {isAuthenticated && user ? (
              <>
                <Link href="/tickets">
                  <Button variant="ghost" size="sm" className="flex items-center gap-2">
                    <Ticket className="h-4 w-4" />
                    <span className="hidden md:inline">Tiket Saya</span>
                  </Button>
                </Link>

                <Link href="/orders">
                  <Button variant="ghost" size="sm" className="flex items-center gap-2">
                    <ShoppingBag className="h-4 w-4" />
                    <span className="hidden md:inline">Pesanan</span>
                  </Button>
                </Link>

                {user.role === "organizer" && (
                  <Link href="/organizer/dashboard">
                    <Button variant="outline" size="sm">
                      Dashboard
                    </Button>
                  </Link>
                )}

                <div className="relative">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setShowUserMenu(!showUserMenu)}
                    className="flex items-center gap-2"
                  >
                    <User className="h-4 w-4" />
                    <span className="hidden md:inline">{user.full_name}</span>
                    <ChevronDown className="h-4 w-4" />
                  </Button>

                  {showUserMenu && (
                    <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border py-1 z-50">
                      <Link href="/profile">
                        <div className="px-4 py-2 hover:bg-gray-100 cursor-pointer flex items-center gap-2">
                          <User className="h-4 w-4" />
                          <span>Profil</span>
                        </div>
                      </Link>
                      <div className="border-t my-1"></div>
                      <div
                        onClick={handleLogout}
                        className="px-4 py-2 hover:bg-gray-100 cursor-pointer flex items-center gap-2 text-red-600"
                      >
                        <LogOut className="h-4 w-4" />
                        <span>Keluar</span>
                      </div>
                    </div>
                  )}
                </div>
              </>
            ) : (
              <>
                <Link href="/login">
                  <Button variant="outline" size="sm">
                    Masuk
                  </Button>
                </Link>
                <Link href="/register">
                  <Button size="sm">Daftar</Button>
                </Link>
              </>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
}
