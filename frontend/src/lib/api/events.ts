import { get } from "./client";
import type { Event, EventDetail, EventsResponse } from "@/types/api";

export interface EventsQueryParams {
  page?: number;
  limit?: number;
  category?: string;
  location?: string;
  search?: string;
  start_date?: string;
  end_date?: string;
  status?: "draft" | "published" | "cancelled";
}

/**
 * Get paginated list of events
 */
export async function getEvents(
  params?: EventsQueryParams
): Promise<EventsResponse> {
  const queryParams = new URLSearchParams();

  if (params?.page) queryParams.set("page", params.page.toString());
  if (params?.limit) queryParams.set("limit", params.limit.toString());
  if (params?.category) queryParams.set("category", params.category);
  if (params?.location) queryParams.set("location", params.location);
  if (params?.search) queryParams.set("search", params.search);
  if (params?.start_date) queryParams.set("start_date", params.start_date);
  if (params?.end_date) queryParams.set("end_date", params.end_date);
  if (params?.status) queryParams.set("status", params.status);

  const query = queryParams.toString();
  const url = query ? `/events?${query}` : "/events";

  return get<EventsResponse>(url);
}

/**
 * Get event by ID with ticket tiers
 */
export async function getEventById(id: string): Promise<EventDetail> {
  return get<EventDetail>(`/events/${id}`);
}

/**
 * Get featured/upcoming events
 */
export async function getFeaturedEvents(): Promise<Event[]> {
  const response = await getEvents({ limit: 6, status: "published" });
  return response.events;
}
