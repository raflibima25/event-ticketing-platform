import Link from "next/link";
import { Card, CardContent } from "@/components/ui/card";
import { formatCurrency, formatDate } from "@/lib/utils";
import { Calendar, MapPin } from "lucide-react";
import type { Event } from "@/types/api";

interface EventCardProps {
  event: Event;
}

export function EventCard({ event }: EventCardProps) {
  return (
    <Link href={`/events/${event.id}`}>
      <Card className="overflow-hidden hover:shadow-lg transition-shadow duration-200 cursor-pointer h-full">
        <div className="aspect-video bg-gradient-to-br from-blue-500 to-purple-600 relative">
          {event.banner_url ? (
            <img
              src={event.banner_url}
              alt={event.title}
              className="w-full h-full object-cover"
            />
          ) : (
            <div className="w-full h-full flex items-center justify-center">
              <Calendar className="h-16 w-16 text-white opacity-50" />
            </div>
          )}
          <div className="absolute top-2 right-2 px-3 py-1 bg-white rounded-full text-xs font-semibold">
            {event.category}
          </div>
        </div>

        <CardContent className="p-4">
          <h3 className="font-semibold text-lg mb-2 line-clamp-2">
            {event.title}
          </h3>

          <div className="space-y-2 text-sm text-gray-600">
            <div className="flex items-start gap-2">
              <Calendar className="h-4 w-4 mt-0.5 flex-shrink-0" />
              <span>{formatDate(event.start_date, "long")}</span>
            </div>

            <div className="flex items-start gap-2">
              <MapPin className="h-4 w-4 mt-0.5 flex-shrink-0" />
              <span>
                {event.venue}, {event.location}
              </span>
            </div>
          </div>

          <div className="mt-4 flex items-center justify-between">
            {event.organizer_name ? (
              <div>
                <p className="text-xs text-gray-500">Diselenggarakan oleh</p>
                <p className="text-sm font-medium">{event.organizer_name}</p>
              </div>
            ) : (
              <div></div>
            )}

            <div className={`px-3 py-1 rounded-full text-xs font-semibold ${
              event.status === "published"
                ? "bg-green-100 text-green-700"
                : event.status === "cancelled"
                ? "bg-red-100 text-red-700"
                : "bg-gray-100 text-gray-700"
            }`}>
              {event.status === "published"
                ? "Tersedia"
                : event.status === "cancelled"
                ? "Dibatalkan"
                : "Draft"}
            </div>
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}
