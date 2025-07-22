import {inject, Injectable} from '@angular/core';
import {ReplaySubject} from "rxjs";
import {ActivatedRoute, Router} from "@angular/router";
import {AuthGuard} from "../_guards/auth.guard";
import {PageService} from "./page.service";

@Injectable({
  providedIn: 'root'
})
export class NavService {

  private pageService = inject(PageService);
  private router = inject(Router);

  private showNavSource = new ReplaySubject<Boolean>(1);
  public showNav$ = this.showNavSource.asObservable();

  private pageIndexSource = new ReplaySubject<number>(1);
  public pageIndex$ = this.pageIndexSource.asObservable();

  constructor(private route: ActivatedRoute) {
    this.showNavSource.next(false);

    this.route.queryParams.subscribe(params => {
      const index = params['index'];
      if (index) {
        try {
          const i = parseInt(index);
          if (i >= 0) {
            this.pageIndexSource.next(i);
          }
        } catch (e) {
          console.error(e);
        }
      }
    })
  }

  setNavVisibility(show: Boolean) {
    this.showNavSource.next(show);
  }

  handleLogin(redirect: boolean = true) {
    this.pageService.refreshPages().subscribe(() => {
      this.setNavVisibility(true);

      if (!redirect) return;

      const pageResume = localStorage.getItem(AuthGuard.urlKey);
      localStorage.setItem(AuthGuard.urlKey, '');
      if (pageResume && pageResume != '/login') {
        this.router.navigateByUrl(pageResume);
      } else {
        this.router.navigateByUrl('/home');
      }
    });
  }
}
