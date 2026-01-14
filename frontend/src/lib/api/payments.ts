import { get, post } from "./client";

export interface Invoice {
  id: string;
  external_id: string;
  user_id: string;
  order_id: string;
  amount: number;
  status: "pending" | "paid" | "expired";
  invoice_url: string;
  expiry_date: string;
  payment_method?: string;
  paid_at?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateInvoiceRequest {
  order_id: string;
  amount: number;
  customer_name: string;
  customer_email: string;
  description: string;
}

/**
 * Create payment invoice for an order
 */
export async function createInvoice(data: CreateInvoiceRequest): Promise<Invoice> {
  return post<Invoice>("/payments/invoices", data);
}

/**
 * Get invoice by ID
 */
export async function getInvoiceById(invoiceId: string): Promise<Invoice> {
  return get<Invoice>(`/payments/invoices/${invoiceId}`);
}

/**
 * Get invoice by order ID
 */
export async function getInvoiceByOrderId(orderId: string): Promise<Invoice> {
  return get<Invoice>(`/payments/orders/${orderId}/invoice`);
}
