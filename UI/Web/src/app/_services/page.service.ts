import {DestroyRef, inject, Injectable} from '@angular/core';
import {HttpClient} from "@angular/common/http";
import {environment} from "../../environments/environment";
import {Page} from "../_models/page";
import {Observable, of, ReplaySubject, tap} from "rxjs";

@Injectable({
  providedIn: 'root'
})
export class PageService {

  private readonly destroyRef = inject(DestroyRef);
  baseUrl = environment.apiUrl + "pages/";

  private pages: Page[] | undefined = undefined;
  private pagesSource = new ReplaySubject<Page[]>(1);
  public pages$ = this.pagesSource.asObservable();

  constructor(private httpClient: HttpClient,) {
    this.refreshPages();
  }

  refreshPages() {
    this.httpClient.get<Page[]>(this.baseUrl).subscribe(pages => {
      this.pages = pages;
      this.pagesSource.next(pages);
    })
  }

  getPage(id: number): Observable<Page> {
    const page = this.pages ? this.pages.find(p => p.id === id) : undefined;
    if (page) {
      return of(page);
    }

    return this.httpClient.get<Page>(this.baseUrl + id)
  }

  removePage(pageId: number) {
    return this.httpClient.delete(this.baseUrl + pageId, {responseType: 'text'});
  }

  upsertPage(page: Page) {
    return this.httpClient.post(this.baseUrl + 'upsert', page, {responseType: 'text'});
  }

  swapPages(id1: number, id2: number) {
    return this.httpClient.post(this.baseUrl + 'swap', {id1, id2}, {responseType: 'text'});
  }

}
