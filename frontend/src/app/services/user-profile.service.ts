import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';
import { getApiErrorInfo } from '@app/shared';
import { environment } from '../../environments/environment';

export interface UserProfile {
  id: string;
  username: string;
  first_name: string;
  last_name: string;
  email: string;
}

export interface UpdateUserProfileDto {
  username: string;
  first_name: string;
  last_name: string;
}

@Injectable({
  providedIn: 'root',
})
export class UserProfileService {
  private readonly http = inject(HttpClient);
  private readonly apiBase = `${environment.apiBaseUrl}/users`;

  async getProfile(): Promise<{ success: boolean; profile?: UserProfile; error?: string }> {
    try {
      const profile = await firstValueFrom(this.http.get<UserProfile>(`${this.apiBase}/profile`));
      return { success: true, profile };
    } catch (error: unknown) {
      const info = getApiErrorInfo(error, 'Failed to load profile');
      return { success: false, error: info.userMessage };
    }
  }

  async updateProfile(
    dto: UpdateUserProfileDto
  ): Promise<{ success: boolean; error?: string; code?: string; fields?: Record<string, string> }> {
    try {
      await firstValueFrom(this.http.patch<void>(`${this.apiBase}/profile`, dto));
      return { success: true };
    } catch (error: unknown) {
      const info = getApiErrorInfo(error, 'Failed to update profile');
      return {
        success: false,
        error: info.userMessage,
        code: info.code,
        fields: info.fields,
      };
    }
  }

  async checkUsernameExists(username: string): Promise<boolean | null> {
    try {
      const result = await firstValueFrom(
        this.http.get<{ exists: boolean }>(
          `${this.apiBase}/check-username?username=${encodeURIComponent(username)}`
        )
      );
      return result.exists;
    } catch {
      return null;
    }
  }
}
