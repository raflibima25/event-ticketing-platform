import { get, post } from "./client";
import type { Order, CreateOrderRequest, CreateOrderResponse, PaginationMeta } from "@/types/api";

/**
 * Create a new order (reserve tickets)
 */
export async function createOrder(data: CreateOrderRequest): Promise<CreateOrderResponse> {
  return post<CreateOrderResponse>("/orders", data);
}

/**
 * Get order by ID
 */
export async function getOrderById(orderId: string): Promise<Order> {
  return get<Order>(`/orders/${orderId}`);
}

/**
 * Get user's orders with pagination
 */
export interface GetOrdersParams {
  page?: number;
  limit?: number;
}

export interface GetOrdersResponse {
  orders: Order[];
  meta: PaginationMeta;
}

export async function getUserOrders(params?: GetOrdersParams): Promise<GetOrdersResponse> {
  const queryParams = new URLSearchParams();

  if (params?.page) queryParams.set("page", params.page.toString());
  if (params?.limit) queryParams.set("limit", params.limit.toString());

  const query = queryParams.toString();
  const url = query ? `/orders?${query}` : "/orders";

  return get<GetOrdersResponse>(url);
}

/**
 * Cancel an order
 */
export async function cancelOrder(orderId: string): Promise<void> {
  return post<void>(`/orders/${orderId}/cancel`, {});
}
