import {DestroyRef, inject, Injectable} from '@angular/core';
import {environment} from "../../environments/environment";
import {HttpClient} from "@angular/common/http";
import {catchError, interval, Observable, ReplaySubject, Subscription} from "rxjs";
import {InfoStat, QueueStat, StatsResponse} from "../_models/stats";
import {DownloadRequest, SearchRequest, StopRequest} from "../_models/search";
import {SearchInfo} from "../_models/Info";

@Injectable({
  providedIn: 'root'
})
export class DownloadService {

  private readonly destroyRef = inject(DestroyRef);
  baseUrl = environment.apiUrl;

  private statsSource = new ReplaySubject<StatsResponse>(1);
  public stats$ = this.statsSource.asObservable();

  private loadStatsSource = new ReplaySubject<Boolean>(1);
  public loadStats$ = this.loadStatsSource.asObservable();

  private sub: Subscription | undefined;
  longSub = false;

  constructor(private httpClient: HttpClient) {
    this.loadStatsSource.next(false)

    this.loadStats$.subscribe(load => {
      this.sub?.unsubscribe();
      if (load) {
        this.sub = interval(1000).subscribe(() => {
          this.refreshStats();
        });
      }
    })


    this.stats$.subscribe(s => {
      if (s.running.length == 0) {
        this.longSub = true;
        this.sub?.unsubscribe();
        this.sub = interval(10000).subscribe(() => {
          this.refreshStats();
        });
      } else if (this.longSub) {
        this.longSub = false;
        this.sub?.unsubscribe();
        this.sub = interval(1000).subscribe(() => {
          this.refreshStats();
        });
      }
    })

  }

  search(req: SearchRequest): Observable<SearchInfo[]> {
    return this.httpClient.post<SearchInfo[]>(this.baseUrl+ 'search', req)
  }

  download(req: DownloadRequest) {
    return this.httpClient.post(this.baseUrl + 'download', req, {responseType: 'text'});
  }

  stop(req: StopRequest) {
    return this.httpClient.post(this.baseUrl + 'stop', req, {responseType: 'text'})
  }

  private refreshStats() {
    this.httpClient.get<StatsResponse>(this.baseUrl + 'stats').subscribe(stats => {
      this.statsSource.next(stats);
    })
  }

  loadStats(load = true) {
    this.loadStatsSource.next(load);
  }
}
