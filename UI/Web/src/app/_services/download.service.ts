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

  private runningSource = new ReplaySubject<InfoStat[] | undefined>(1)
  public running$ = this.runningSource.asObservable();

  private queuedSource = new ReplaySubject<QueueStat[] | undefined>(1)
  public queued$ = this.queuedSource.asObservable();

  private loadStatsSource = new ReplaySubject<Boolean>(1);
  public loadStats$ = this.loadStatsSource.asObservable();

  private sub: Subscription | undefined;

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

    this.running$.subscribe(running => {
      if (running?.length == 0) {
        this.sub?.unsubscribe();
        this.sub = interval(10000).subscribe(() => {
          this.refreshStats();
        });
      }
    })

  }

  search(req: SearchRequest): Observable<SearchInfo[]> {
    return this.httpClient.post<SearchInfo[]>(this.baseUrl+ 'search', req)
  }

  download(req: DownloadRequest) {
    console.log(this.baseUrl + 'download')
    return this.httpClient.post(this.baseUrl + 'download', req);
  }

  stop(req: StopRequest) {
    return this.httpClient.post(this.baseUrl + 'stop', stop)
  }

  private refreshStats() {
    this.httpClient.get<StatsResponse>(this.baseUrl + 'stats').subscribe(stats => {
      this.runningSource.next(stats.running);
      this.queuedSource.next(stats.queued);
    })
  }

  loadStats(load = true) {
    this.loadStatsSource.next(load);
  }
}