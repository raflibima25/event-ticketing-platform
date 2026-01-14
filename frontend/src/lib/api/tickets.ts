import { get } from "./client";
import type { Ticket } from "@/types/api";

/**
 * Get ticket by ID
 */
export async function getTicketById(ticketId: string): Promise<Ticket> {
  return get<Ticket>(`/tickets/${ticketId}`);
}

/**
 * Get user's tickets
 */
export async function getUserTickets(): Promise<Ticket[]> {
  return get<Ticket[]>("/tickets");
}
