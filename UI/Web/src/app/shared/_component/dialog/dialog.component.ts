import {Component, HostListener, Input, OnInit, Renderer2} from '@angular/core';
import {ReplaySubject} from "rxjs";
import {NgClass} from "@angular/common";

@Component({
  selector: 'app-dialog',
  standalone: true,
  imports: [
    NgClass
  ],
  templateUrl: './dialog.component.html',
  styleUrl: './dialog.component.css'
})
export class DialogComponent implements OnInit {

  @Input() isMobile = false;
  @Input() text: string = '';

  private result = new ReplaySubject<boolean>(1)



  @HostListener('window:resize', ['$event'])
  onResize() {
    this.isMobile = window.innerWidth < 768;
  }

  ngOnInit(): void {
    this.isMobile = window.innerWidth < 768;
  }

  public getResult() {
    return this.result.asObservable();
  }

  closeDialog() {
    this.result.next(false);
    this.result.complete();
  }

  confirm() {
    this.result.next(true);
    this.result.complete();
  }

}
