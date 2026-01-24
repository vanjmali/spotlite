import { inject, Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';

export interface Genre {
  id: string;
  name: string;
}

export interface CreateGenreDto {
  name: string;
}

export interface UpdateGenreDto {
  name?: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
}

@Injectable({
  providedIn: 'root',
})
export class GenreService {
  private http = inject(HttpClient);
  private apiUrl = '/api/content/genres';

  getGenres(
    page: number = 1,
    size: number = 50,
    filters?: {
      name?: string;
    }
  ): Observable<PaginatedResponse<Genre>> {
    let params = new HttpParams().set('page', page.toString()).set('size', size.toString());

    if (filters?.name) {
      params = params.set('name', filters.name);
    }

    return this.http.get<PaginatedResponse<Genre>>(this.apiUrl, { params });
  }

  getGenreById(id: string): Observable<Genre> {
    return this.http.get<Genre>(`${this.apiUrl}/${id}`);
  }

  createGenre(dto: CreateGenreDto): Observable<void> {
    return this.http.post<void>(this.apiUrl, dto);
  }

  updateGenre(id: string, dto: UpdateGenreDto): Observable<Genre> {
    return this.http.patch<Genre>(`${this.apiUrl}/${id}`, dto);
  }

  deleteGenre(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`);
  }
}
