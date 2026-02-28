import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, map } from 'rxjs';
import { Album } from './album.service';
import { Artist } from './artist.service';
import { Song } from './song.service';
import { Genre } from './genre.service';

export interface GlobalSearchResponse {
  songs: Song[];
  artists: Artist[];
  albums: Album[];
  genres: Genre[];
}

export type SearchEntityType = 'song' | 'artist' | 'album' | 'genre';

export interface SearchSuggestion {
  id: string;
  type: SearchEntityType;
  label: string;
  subtitle: string;
}

@Injectable({
  providedIn: 'root',
})
export class ContentSearchService {
  private readonly http = inject(HttpClient);
  private readonly apiUrl = '/api/content/search';

  search(query: string): Observable<GlobalSearchResponse> {
    const params = new HttpParams().set('q', query);
    return this.http.get<GlobalSearchResponse>(this.apiUrl, { params });
  }

  topSuggestions(query: string, max: number = 5): Observable<SearchSuggestion[]> {
    return this.search(query).pipe(
      map((result) => {
        const normalized = query.trim().toLowerCase();
        const suggestions: Array<SearchSuggestion & { score: number }> = [];

        for (const song of result.songs ?? []) {
          suggestions.push({
            id: song.id,
            type: 'song',
            label: song.title,
            subtitle: song.artists?.map((artist) => artist.name).join(', ') || 'Song',
            score: rankSuggestion(song.title, normalized),
          });
        }

        for (const artist of result.artists ?? []) {
          suggestions.push({
            id: artist.id,
            type: 'artist',
            label: artist.name,
            subtitle: 'Artist',
            score: rankSuggestion(artist.name, normalized),
          });
        }

        for (const album of result.albums ?? []) {
          suggestions.push({
            id: album.id,
            type: 'album',
            label: album.title,
            subtitle: album.artists?.map((artist) => artist.name).join(', ') || 'Album',
            score: rankSuggestion(album.title, normalized),
          });
        }

        for (const genre of result.genres ?? []) {
          suggestions.push({
            id: genre.id,
            type: 'genre',
            label: genre.name,
            subtitle: 'Genre',
            score: rankSuggestion(genre.name, normalized),
          });
        }

        return suggestions
          .sort((a, b) => b.score - a.score || a.label.localeCompare(b.label))
          .slice(0, max)
          .map((suggestion) => ({
            id: suggestion.id,
            type: suggestion.type,
            label: suggestion.label,
            subtitle: suggestion.subtitle,
          }));
      })
    );
  }
}

function rankSuggestion(text: string, normalizedQuery: string): number {
  const normalizedText = text.trim().toLowerCase();
  if (!normalizedText || !normalizedQuery) {
    return 0;
  }

  if (normalizedText === normalizedQuery) {
    return 100;
  }
  if (normalizedText.startsWith(normalizedQuery)) {
    return 80;
  }
  if (normalizedText.includes(normalizedQuery)) {
    return 60;
  }

  return 1;
}
