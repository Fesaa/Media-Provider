import {DestroyRef, inject, Injectable} from '@angular/core';
import {HttpClient} from "@angular/common/http";
import {environment} from "../../environments/environment";
import {DownloadMetadata, Page, Provider} from "../_models/page";
import {Observable, of, ReplaySubject} from "rxjs";

@Injectable({
  providedIn: 'root'
})
export class PageService {

  public static readonly DEFAULT_PAGE_SORT = 9999;

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
    const page = this.pages ? this.pages.find(p => p.ID === id) : undefined;
    if (page) {
      return of(page);
    }

    return this.httpClient.get<Page>(this.baseUrl + id)
  }

  removePage(pageId: number) {
    return this.httpClient.delete(this.baseUrl + pageId);
  }

  new(page: Page) {
    return this.httpClient.post<Page>(this.baseUrl + 'new', page);
  }

  update(page: Page) {
    return this.httpClient.post<Page>(this.baseUrl + 'update', page);
  }

  swapPages(id1: number, id2: number) {
    return this.httpClient.post(this.baseUrl + 'swap', {id1, id2});
  }

  loadDefault() {
    if (this.pages != undefined && this.pages.length !== 0) {
      throw "Cannot load default while pages are available"
    }

    return this.httpClient.post(this.baseUrl + "load-default", {})
  }

  metadata(provider: Provider) {
    return this.httpClient.get<DownloadMetadata>(this.baseUrl + `download-metadata?provider=${provider}`)
  }

}
