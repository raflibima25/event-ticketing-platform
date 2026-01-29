// ============================================
// API Response Types
// ============================================

export interface ApiResponse<T = unknown> {
  status: boolean;
  message: string;
  data?: T;
}

export interface ApiError {
  status: boolean;
  message: string;
  errors?: any;
  error_code?: string;
  timestamp?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    current_page: number;
    per_page: number;
    total: number;
    total_pages: number;
  };
}

export interface PaginationMeta {
  current_page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

export interface EventsResponse {
  events: Event[];
  meta: PaginationMeta;
}

// ============================================
// Auth Types
// ============================================

export interface User {
  id: string;
  email: string;
  full_name: string;
  phone?: string;
  role: "customer" | "organizer" | "admin";
  is_email_verified?: boolean;
  created_at: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  full_name: string;
  phone: string;
  role?: "customer" | "organizer";
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ResetPasswordRequest {
  token: string;
  new_password: string;
}

// ============================================
// Event Types
// ============================================

export interface Event {
  id: string;
  title: string;
  slug: string;
  description: string;
  category: string;
  location: string;
  venue: string;
  start_date: string;
  end_date: string;
  timezone: string;
  banner_url?: string;
  organizer_id: string;
  organizer_name?: string;
  status: "draft" | "published" | "cancelled";
  created_at: string;
  updated_at: string;
}

export interface TicketTier {
  id: string;
  event_id: string;
  name: string;
  description?: string;
  price: number;
  quota: number;
  sold_count: number;
  available_count: number;
  max_per_order: number;
  early_bird_price?: number;
  early_bird_end_date?: string;
  current_price: number;
  is_sold_out: boolean;
  created_at: string;
  updated_at: string;
}

export interface EventDetail extends Event {
  ticket_tiers: TicketTier[];
}

// ============================================
// Order Types
// ============================================

export interface OrderItem {
  id: string;
  ticket_tier_id: string;
  quantity: number;
  price: number;
  subtotal: number;
}

export interface Order {
  id: string;
  user_id: string;
  event_id: string;
  items: OrderItem[];
  total_amount: number;
  platform_fee: number;
  service_fee: number;
  grand_total: number;
  status: "reserved" | "paid" | "cancelled" | "expired";
  payment_id?: string;
  payment_method?: string;
  reservation_expires_at: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
}

export interface CreateOrderRequest {
  event_id: string;
  items: {
    ticket_tier_id: string;
    quantity: number;
  }[];
}

export interface CreateOrderResponse {
  order: Order;
  invoice_url: string;
}

// ============================================
// Ticket Types
// ============================================

export interface Ticket {
  id: string;
  order_id: string;
  order_item_id: string;
  ticket_tier_id: string;
  ticket_tier_name: string;
  event_id: string;
  event_title: string;
  event_date: string;
  event_location: string;
  user_id: string;
  user_name: string;
  user_email: string;
  qr_code: string;
  status: "active" | "used" | "cancelled";
  validated_at?: string;
  created_at: string;
}
