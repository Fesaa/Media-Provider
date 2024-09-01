import {DestroyRef, inject, Injectable} from '@angular/core';
import {HttpClient} from "@angular/common/http";
import {environment} from "../../environments/environment";
import {Page} from "../_models/page";
import {Observable, of, tap} from "rxjs";
import {takeUntilDestroyed} from "@angular/core/rxjs-interop";

@Injectable({
  providedIn: 'root'
})
export class PageService {

  private readonly destroyRef = inject(DestroyRef);
  baseUrl = environment.apiUrl;

  private pages: Page[] | undefined = undefined;
  public currentPage: number | undefined = undefined;

  constructor(private httpClient: HttpClient) { }

  getPages(): Observable<Page[]> {
    if (this.pages !== undefined) {
      return of(this.pages);
    }

    return this.httpClient.get<Page[]>(this.baseUrl + 'pages/').pipe(
      tap(pages => {
        this.pages = pages;
      }), takeUntilDestroyed(this.destroyRef)
    );
  }

  getPage(index: number): Observable<Page> {
    if (this.pages && index < this.pages.length && index > 0) {
      return of(this.pages[index]);
    }

    return this.httpClient.get<Page>(this.baseUrl + 'pages/' + index)
  }
}
