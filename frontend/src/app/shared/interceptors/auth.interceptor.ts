import { Injectable, inject } from '@angular/core';
import {
  HttpRequest,
  HttpHandler,
  HttpEvent,
  HttpInterceptor,
  HttpErrorResponse,
} from '@angular/common/http';
import { Observable, throwError, from } from 'rxjs';
import { catchError, switchMap } from 'rxjs/operators';
import { AuthService } from '../../services/auth.service';
import { Router } from '@angular/router';

/**
 * HTTP Interceptor that:
 * 1. Injects JWT access token into all API requests (except public endpoints)
 * 2. Handles 401 Unauthorized responses by attempting to refresh token
 * 3. Logs user out on refresh failure
 */
@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  private authService = inject(AuthService);
  private router = inject(Router);

  intercept(request: HttpRequest<unknown>, next: HttpHandler): Observable<HttpEvent<unknown>> {
    // Get current access token
    const token = this.authService.accessTokenSg();

    // Add JWT token to request if available and not a public/auth request
    const excludedPaths = ['/register', '/login', '/refresh-token', '/logout'];
    const isExcluded = excludedPaths.some((path) => request.url.includes(path));
    if (token && !isExcluded) {
      request = request.clone({
        setHeaders: {
          Authorization: `Bearer ${token}`,
        },
      });
    }

    return next.handle(request).pipe(
      catchError((error: HttpErrorResponse) => {
        // Handle 401 Unauthorized - try to refresh token
        if (error.status === 401) {
          return from(this.authService.refreshAccessToken()).pipe(
            switchMap((success) => {
              if (success) {
                // Retry request with new token
                const newToken = this.authService.accessTokenSg();
                const retryReq = request.clone({
                  setHeaders: {
                    Authorization: `Bearer ${newToken}`,
                  },
                });
                return next.handle(retryReq);
              } else {
                // Refresh failed, logout user
                this.authService.logout();
                if (!this.router.url.includes('/login')) {
                  this.router.navigate(['/login']);
                }
                return throwError(() => error);
              }
            }),
            catchError(() => {
              this.authService.logout();
              if (!this.router.url.includes('/login')) {
                this.router.navigate(['/login']);
              }
              return throwError(() => error);
            })
          );
        }

        return throwError(() => error);
      })
    );
  }
}
