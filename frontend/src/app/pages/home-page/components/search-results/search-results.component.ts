import { CommonModule } from '@angular/common';
import { Component, DestroyRef, inject, signal } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { ContentSearchService, GlobalSearchResponse } from '@app/services/content-search.service';
import { Artist } from '@app/services/artist.service';
import { debounceTime, distinctUntilChanged, map, of, switchMap, catchError } from 'rxjs';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';
import { SongRatingBadgeComponent } from '@app/shared/components/song-rating-badge/song-rating-badge';

@Component({
  selector: 'app-search-results',
  standalone: true,
  imports: [CommonModule, RouterLink, CoverArtComponent, SongRatingBadgeComponent],
  templateUrl: './search-results.component.html',
  styleUrl: './search-results.component.scss',
})
export class SearchResultsComponent {
  private readonly route = inject(ActivatedRoute);
  private readonly destroyRef = inject(DestroyRef);
  private readonly searchService = inject(ContentSearchService);

  readonly querySg = signal('');
  readonly isLoadingSg = signal(false);
  readonly resultsSg = signal<GlobalSearchResponse>({
    songs: [],
    artists: [],
    albums: [],
    genres: [],
  });

  constructor() {
    this.route.queryParamMap
      .pipe(
        map((params) => params.get('q')?.trim() ?? ''),
        debounceTime(150),
        distinctUntilChanged(),
        switchMap((query) => {
          this.querySg.set(query);
          if (query.length < 2) {
            this.resultsSg.set({ songs: [], artists: [], albums: [], genres: [] });
            return of(null);
          }

          this.isLoadingSg.set(true);
          return this.searchService.search(query).pipe(
            catchError(() =>
              of<GlobalSearchResponse>({
                songs: [],
                artists: [],
                albums: [],
                genres: [],
              })
            )
          );
        }),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe((result) => {
        if (result) {
          this.resultsSg.set(result);
        }
        this.isLoadingSg.set(false);
      });
  }

  artistNames(artists: Artist[] | undefined): string {
    if (!artists || artists.length === 0) {
      return 'Unknown artist';
    }

    return artists.map((artist) => artist.name).join(', ');
  }
}
