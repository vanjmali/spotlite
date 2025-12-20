import { Injectable, signal } from '@angular/core';

export interface MockUser {
  id: string;
  firstName: string;
  lastName: string;
  email: string;
  username: string;
  passwordHash: string;
}

export type OtpVerificationStatus = 'success' | 'invalid' | 'expired';

@Injectable({ providedIn: 'root' })
export class AuthService {
  // signals to hold auth state
  currentEmail = signal<string | null>(null);
  isAuthenticated = signal(false);

  // TODO: Replace mock users with real backend authentication
  private users: MockUser[] = [];

  // simulate checking email exists
  async checkEmail(email: string): Promise<{ exists: boolean; user?: MockUser }> {
    await this.delay(300);
    const user = this.users.find(
      (u) =>
        u.email.toLowerCase() === email.toLowerCase() ||
        u.username.toLowerCase() === email.toLowerCase()
    );
    return { exists: !!user, user };
  }

  // TODO: Replace with real email OTP service
  // simulate sending a 6-digit OTP to email (mock)
  private otpStore = new Map<string, { code: string; expiresAt: number }>();
  private readonly OTP_EXPIRY_MS = 5 * 60 * 1000; // 5 minutes

  async sendOtp(email: string): Promise<void> {
    await this.delay(200);
    const code = Math.floor(100000 + Math.random() * 900000).toString();
    this.otpStore.set(email.toLowerCase(), {
      code,
      expiresAt: Date.now() + this.OTP_EXPIRY_MS,
    });
    // TODO: Send OTP via email service
  }

  async verifyOtp(email: string, code: string): Promise<OtpVerificationStatus> {
    await this.delay(200);
    const otp = this.otpStore.get(email.toLowerCase());

    if (!otp) {
      return 'invalid';
    }

    // Check if OTP has expired
    if (Date.now() > otp.expiresAt) {
      this.otpStore.delete(email.toLowerCase());
      return 'expired';
    }

    if (otp.code === code) {
      this.isAuthenticated.set(true);
      this.currentEmail.set(email);
      this.otpStore.delete(email.toLowerCase());
      return 'success';
    }
    return 'invalid';
  }

  // TODO: Replace with real password-based login against backend
  async login(email: string, password: string): Promise<boolean> {
    await this.delay(400);
    const found = this.users.find(
      (u) =>
        (u.email.toLowerCase() === email.toLowerCase() ||
          u.username.toLowerCase() === email.toLowerCase()) &&
        u.passwordHash === this.hashPassword(password)
    );
    if (found) {
      this.currentEmail.set(found.email);
      this.isAuthenticated.set(true);
      return true;
    }
    this.isAuthenticated.set(false);
    return false;
  }

  // Check if email already exists
  async checkEmailExists(email: string): Promise<boolean> {
    await this.delay(300);
    return this.users.some((u) => u.email.toLowerCase() === email.toLowerCase());
  }

  // Check if username already exists
  async checkUsernameExists(username: string): Promise<boolean> {
    await this.delay(300);
    return this.users.some((u) => u.username.toLowerCase() === username.toLowerCase());
  }

  // Register new user (add to mock users)
  async register(
    firstName: string,
    lastName: string,
    email: string,
    username: string,
    password: string
  ): Promise<boolean> {
    await this.delay(400);

    // Check if email exists
    if (await this.checkEmailExists(email)) {
      return false;
    }

    // Check if username exists
    if (await this.checkUsernameExists(username)) {
      return false;
    }

    // Add new user to mock users
    const newUser: MockUser = {
      id: (this.users.length + 1).toString(),
      firstName,
      lastName,
      email,
      username,
      passwordHash: this.hashPassword(password),
    };

    this.users.push(newUser);
    return true;
  }

  private delay(ms: number) {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  private hashPassword(value: string): string {
    // Lightweight hash to avoid storing mock passwords in plain text
    return btoa(value);
  }
}
