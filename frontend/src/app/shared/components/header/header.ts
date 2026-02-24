import { CommonModule } from '@angular/common';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '@app/services/auth.service';
import { UserProfileDropdownComponent } from './components';
import { Component, DestroyRef, inject, input, signal } from '@angular/core';
import { ReactiveFormsModule, FormControl } from '@angular/forms';
import { ContentSearchService, SearchSuggestion } from '@app/services/content-search.service';
import { debounceTime, distinctUntilChanged, of, switchMap, catchError, tap } from 'rxjs';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { MatIconModule } from '@angular/material/icon';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    ReactiveFormsModule,
    UserProfileDropdownComponent,
    MatIconModule,
  ],
  templateUrl: './header.html',
  styleUrls: ['./header.scss'],
})
export class HeaderComponent {
  private readonly router = inject(Router);
  private readonly destroyRef = inject(DestroyRef);
  private readonly searchService = inject(ContentSearchService);
  readonly authService = inject(AuthService);
  readonly presentLogoSg = input<boolean>(false);
  readonly searchControl = new FormControl('', { nonNullable: true });
  readonly suggestionsSg = signal<SearchSuggestion[]>([]);
  readonly suggestionsOpenSg = signal(false);
  readonly isSearchingSg = signal(false);

  navigateToHome(): void {
    this.router.navigate(['/']);
  }

  constructor() {
    this.searchControl.valueChanges
      .pipe(
        debounceTime(3000),
        distinctUntilChanged(),
        switchMap((query) => {
          const normalized = query.trim();
          if (normalized.length < 2) {
            this.isSearchingSg.set(false);
            return of([]);
          }

          this.isSearchingSg.set(true);
          return this.searchService.topSuggestions(normalized, 5).pipe(
            catchError(() => of([])),
            tap(() => this.isSearchingSg.set(false))
          );
        }),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe((suggestions) => {
        this.suggestionsSg.set(suggestions);
        this.suggestionsOpenSg.set(suggestions.length > 0);
      });
  }

  onSearchFocus(): void {
    if (this.suggestionsSg().length > 0) {
      this.suggestionsOpenSg.set(true);
    }
  }

  onSearchBlur(): void {
    setTimeout(() => this.suggestionsOpenSg.set(false), 120);
  }

  openSuggestion(suggestion: SearchSuggestion): void {
    this.suggestionsOpenSg.set(false);
    this.routeForType(suggestion.type, suggestion.id);
  }

  submitSearch(): void {
    const value = this.searchControl.value.trim();
    if (!value) {
      return;
    }

    this.suggestionsOpenSg.set(false);
    this.router.navigate(['/search'], { queryParams: { q: value } });
  }

  labelForType(type: SearchSuggestion['type']): string {
    switch (type) {
      case 'song':
        return 'Song';
      case 'artist':
        return 'Artist';
      case 'album':
        return 'Album';
      case 'genre':
        return 'Genre';
      default:
        return 'Item';
    }
  }

  private routeForType(type: SearchSuggestion['type'], id: string): void {
    switch (type) {
      case 'song':
        this.router.navigate(['/search'], { queryParams: { q: this.searchControl.value.trim() } });
        break;
      case 'artist':
        this.router.navigate(['/artist', id]);
        break;
      case 'album':
        this.router.navigate(['/album', id]);
        break;
      case 'genre':
        this.router.navigate(['/search'], { queryParams: { q: this.searchControl.value.trim() } });
        break;
      default:
        this.router.navigate(['/search'], { queryParams: { q: this.searchControl.value.trim() } });
    }
  }
}
