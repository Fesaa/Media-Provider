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
  baseUrl = environment.apiUrl;

  private pages: Page[] | undefined = undefined;
  private pagesSource = new ReplaySubject<Page[]>(1);
  public pages$ = this.pagesSource.asObservable();

  constructor(private httpClient: HttpClient,) {
    this.refreshPages();
  }

  refreshPages() {
    this.httpClient.get<Page[]>(this.baseUrl + 'config/pages/').subscribe(pages => {
      this.pages = pages;
      this.pagesSource.next(pages);
    })
  }

  getPage(index: number): Observable<Page> {
    if (this.pages && index < this.pages.length && index >= 0) {
      return of(this.pages[index]);
    }

    return this.httpClient.get<Page>(this.baseUrl + 'config/pages/' + index)
  }
}
