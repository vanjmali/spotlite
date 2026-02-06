import { Injectable, inject } from '@angular/core';
import {
  HttpRequest,
  HttpHandler,
  HttpEvent,
  HttpInterceptor,
  HttpErrorResponse,
} from '@angular/common/http';
import { Observable, throwError, from, BehaviorSubject } from 'rxjs';
import { catchError, filter, switchMap, take } from 'rxjs/operators';
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
  private refreshTokenInProgress = false;
  private refreshTokenSubject = new BehaviorSubject<string | null>(null);

  intercept(request: HttpRequest<unknown>, next: HttpHandler): Observable<HttpEvent<unknown>> {
    // get the access token stored in memory
    const token = this.authService.accessTokenSg();

    // exclude auth-related endpoints
    const excludedPaths = ['/register', '/login', '/refresh-token', '/logout'];
    const isExcluded = excludedPaths.some((path) => request.url.includes(path));

    // add token to non-excluded requests
    if (token && !isExcluded) {
      request = this.addToken(request, token);
    }

    // send the request with the access token attached
    return next.handle(request).pipe(
      catchError((error: HttpErrorResponse) => {
        // only handle 401 on non-excluded paths
        if (error.status === 401 && !isExcluded) {
          return this.handle401Error(request, next);
        }

        // if any other type of error occurs, throw it
        return throwError(() => error);
      })
    );
  }

  // used to clone a request and adds the access token in the authorization headers
  private addToken(request: HttpRequest<unknown>, token: string): HttpRequest<unknown> {
    return request.clone({
      setHeaders: {
        Authorization: `Bearer ${token}`,
      },
    });
  }

  // used to handle 401 errors (attempt token refreshing, retry request with new token/log out)
  private handle401Error(
    request: HttpRequest<unknown>,
    next: HttpHandler
  ): Observable<HttpEvent<unknown>> {
    // boolean lock to make sure only one refresh is happening at a moment
    if (!this.refreshTokenInProgress) {
      this.refreshTokenInProgress = true;

      // reset the subject to signal a new refresh is starting
      this.refreshTokenSubject.next(null);

      // refresh the token, then use switchMap to transform the result into a retry of the original request
      return from(this.authService.refreshAccessToken()).pipe(
        switchMap((success) => {
          this.refreshTokenInProgress = false;

          // if the token is refreshed successfully store it and update the refreshTokenSubject
          if (success) {
            const newToken = this.authService.accessTokenSg();
            this.refreshTokenSubject.next(newToken);

            // retry the request with the new token
            return next.handle(this.addToken(request, newToken!)).pipe(
              catchError((retryError: HttpErrorResponse) => {
                // logout if retry also gets 401
                if (retryError.status === 401) {
                  this.handleLogout();
                }

                return throwError(() => retryError);
              })
            );
          } else {
            this.handleLogout();
            return throwError(() => new Error('Token refresh failed'));
          }
        }),
        // this catch handles errors from both the refresh AND the retry
        catchError((error) => {
          this.refreshTokenInProgress = false;

          // only logout if the refresh endpoint itself failed, not if the retried request failed
          if (error instanceof HttpErrorResponse && error.url?.includes('/refresh-token')) {
            this.handleLogout();
          }

          return throwError(() => error);
        })
      );
    } else {
      // if the token refresh is already in progress just wait until it's done, take the first result
      // and retry the request
      return this.refreshTokenSubject.pipe(
        filter((token) => token !== null),
        take(1),
        switchMap((token) => {
          return next.handle(this.addToken(request, token!)).pipe(
            catchError((error: HttpErrorResponse) => {
              return throwError(() => error);
            })
          );
        })
      );
    }
  }

  private handleLogout(): void {
    this.authService.logout();
    if (!this.router.url.includes('/login')) {
      this.router.navigate(['/']);
    }
  }
}
