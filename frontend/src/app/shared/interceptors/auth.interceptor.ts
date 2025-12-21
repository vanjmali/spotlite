import { Injectable, inject } from '@angular/core';
import {
  HttpRequest,
  HttpHandler,
  HttpEvent,
  HttpInterceptor,
  HttpErrorResponse,
} from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { AuthService } from '../../services/auth.service';
import { Router } from '@angular/router';

/**
 * HTTP Interceptor that:
 * 1. Injects JWT access token into all API requests (except public endpoints)
 * 2. Handles 401 Unauthorized responses by clearing auth state
 * 3. Logs errors for debugging
 */
@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  private authService = inject(AuthService);
  private router = inject(Router);

  intercept(request: HttpRequest<unknown>, next: HttpHandler): Observable<HttpEvent<unknown>> {
    // Get current access token
    const token = this.authService.accessTokenSg();

    // Add JWT token to request if available and not a login/register request
    if (token && !request.url.includes('/register') && !request.url.includes('/login')) {
      request = request.clone({
        setHeaders: {
          Authorization: `Bearer ${token}`,
        },
      });
    }

    return next.handle(request).pipe(
      catchError((error: HttpErrorResponse) => {
        // Handle 401 Unauthorized - token expired or invalid
        if (error.status === 401) {
          // Clear auth state
          this.authService.logout();

          // Redirect to login if not already there
          if (!this.router.url.includes('/login')) {
            this.router.navigate(['/login']);
          }
        }

        // Log error
        console.error('HTTP Error:', error);

        return throwError(() => error);
      })
    );
  }
}
