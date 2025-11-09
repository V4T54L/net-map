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

// Decoded JWT payload from backend
export interface DecodedToken {
  UserID: number;
  Role: 'user' | 'admin';
  exp: number;
  iat: number;
  sub: string; // Subject, which is the username
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
```
