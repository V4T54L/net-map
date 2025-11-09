export interface AuthTokens {
  AccessToken: string;
  RefreshToken: string;
}

export interface User {
  ID: number;
  Username: string;
  Role: 'user' | 'admin';
  IsEnabled: boolean;
}

export interface DecodedToken {
  UserID: number;
  Role: 'user' | 'admin';
  exp: number;
  iat: number;
  sub: string;
}

export interface AuthContextType {
  user: User | null;
  tokens: AuthTokens | null;
  loading: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  register: (username: string, password: string) => Promise<void>;
}

export interface RegisterRequest {
  Username: string;
  Password: string;
}

export interface LoginRequest {
  Username: string;
  Password: string;
}

export interface AuthResponse {
  AccessToken: string;
  RefreshToken: string;
}

export interface DNSRecord {
  ID: number;
  UserID: number;
  DomainName: string;
  Type: 'A' | 'CNAME';
  Value: string;
  CreatedAt: string;
  UpdatedAt: string;
}

export interface CreateDNSRecordRequest {
  DomainName: string;
  Type: 'A' | 'CNAME';
  Value: string;
}

export interface UpdateDNSRecordRequest {
  DomainName?: string;
  Type?: 'A' | 'CNAME';
  Value?: string;
}
