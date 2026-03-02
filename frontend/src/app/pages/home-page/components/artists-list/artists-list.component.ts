import { CommonModule } from '@angular/common';
import { Component, inject, signal } from '@angular/core';
import { forkJoin, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import {
  RecommendationService,
  type SongRecommendation,
} from '@app/services/recommendation.service';
import { CoverArtComponent } from '@app/shared/components/cover-art/cover-art';
import { MessageComponent } from '@app/shared/components/message';
import { SongTrailingMetaComponent } from '@app/shared/components/song-trailing-meta/song-trailing-meta';

@Component({
  selector: 'app-artists-list',
  standalone: true,
  imports: [CommonModule, CoverArtComponent, MessageComponent, SongTrailingMetaComponent],
  templateUrl: './artists-list.component.html',
  styleUrl: './artists-list.component.scss',
})
export class ArtistsListComponent {
  private readonly recommendationService = inject(RecommendationService);

  readonly isLoadingSg = signal(true);
  readonly subscriptionRecommendationsSg = signal<SongRecommendation[]>([]);
  readonly likeRecommendationsSg = signal<SongRecommendation[]>([]);
  readonly subscriptionErrorSg = signal<string>('');
  readonly likeErrorSg = signal<string>('');

  constructor() {
    this.loadRecommendations();
  }

  private loadRecommendations(): void {
    this.isLoadingSg.set(true);
    this.subscriptionErrorSg.set('');
    this.likeErrorSg.set('');

    forkJoin({
      subscription: this.recommendationService.getSubscriptionRecommendations().pipe(
        catchError(() => {
          this.subscriptionErrorSg.set('Failed to load recommendation songs for your subscriptions.');
          return of([]);
        })
      ),
      likes: this.recommendationService.getLikeBasedRecommendations().pipe(
        catchError(() => {
          this.likeErrorSg.set('Failed to load top liked recommendation songs.');
          return of([]);
        })
      ),
    }).subscribe({
      next: (result) => {
        this.subscriptionRecommendationsSg.set(result.subscription ?? []);
        this.likeRecommendationsSg.set(result.likes ?? []);
        this.isLoadingSg.set(false);
      },
      error: () => {
        this.subscriptionErrorSg.set('Failed to load recommendation songs.');
        this.likeErrorSg.set('Failed to load recommendation songs.');
        this.isLoadingSg.set(false);
      },
    });
  }

  artistNames(artists: string[] | undefined): string {
    if (!artists || artists.length === 0) {
      return 'Unknown artist';
    }

    return artists.join(', ');
  }
}
