import {Component, EventEmitter, HostListener, Input, OnInit, Output} from '@angular/core';
import {ReplaySubject} from "rxjs";
import {NgIcon} from "@ng-icons/core";
import {DirEntry} from "../../../_models/io";
import {Stack} from "../../data-structures/stack";
import {IoService} from "../../../_services/io.service";
import {Clipboard} from "@angular/cdk/clipboard";
import {FormsModule} from "@angular/forms";
import {Dialog} from "primeng/dialog";
import {Button} from "primeng/button";
import {MessageService} from "../../../_services/message.service";

@Component({
  selector: 'app-directory-selector',
  imports: [
    NgIcon,
    FormsModule,
    Dialog,
    Button
  ],
  templateUrl: './directory-selector.component.html',
  styleUrl: './directory-selector.component.css'
})
export class DirectorySelectorComponent implements OnInit {

  @Input() isMobile = false;

  @Input({required: true}) root!: string;
  @Input() showFiles: boolean = false;
  @Input() filter: boolean = false;
  @Input() copy: boolean = true;
  @Input() create: boolean = false;
  @Input() customWidth: string = '50vw';

  @Input() visible: boolean = true;
  @Output() visibleChange: EventEmitter<boolean> = new EventEmitter<boolean>();

  @Output() resultDir = new EventEmitter<string | undefined>();

  currentRoot = '';
  entries: DirEntry[] = [];
  routeStack: Stack<string> = new Stack<string>();

  query: string = '';
  newDirName: string = '';
  private result = new ReplaySubject<string | undefined>(1)

  constructor(private ioService: IoService,
              private msgService: MessageService,
              private clipboard: Clipboard,
  ) {
  }

  ngOnInit(): void {
    this.currentRoot = this.root;
    this.routeStack.push(this.root);
    this.loadChildren(this.root);
    this.isMobile = window.innerWidth < 768;
  }

  getEntries() {
    return this.entries.filter(entry => this.normalize(entry.name).includes(this.query));
  }

  selectNode(entry: DirEntry) {
    if (!entry.dir) {
      return;
    }

    this.query = '';
    this.currentRoot = entry.name;
    this.routeStack.push(entry.name);
    this.loadChildren(this.routeStack.items.join('/'));
  }

  goBack() {
    if (this.routeStack.isEmpty()) {
      return;
    }

    this.routeStack.pop();
    const nextRoot = this.routeStack.peek();
    if (nextRoot) {
      this.currentRoot = nextRoot;
    }
    this.loadChildren(this.routeStack.items.join('/'));
  }

  normalize(str: string): string {
    return str.toLowerCase();
  }

  onFilterChange(event: Event) {
    const inputElement = event.target as HTMLInputElement;
    this.query = this.normalize(inputElement.value);
  }

  onNewDirNameChange(event: Event) {
    const inputElement = event.target as HTMLInputElement;
    this.newDirName = inputElement.value;
  }

  createNew() {
    this.ioService.create(this.routeStack.items.join('/'), this.newDirName).subscribe({
      next: () => {
        this.msgService.success('Success', `Directory ${this.newDirName} created successfully`);
        this.newDirName = '';
        this.loadChildren(this.routeStack.items.join('/'));
      },
      error: (err) => {
        this.msgService.error('Error', `Failed to create directory ${this.newDirName}\n${err.error.message}`);
        console.error(err);
      }
    });
  }

  copyPath(entry: DirEntry) {
    let path = this.routeStack.items.join('/');
    if (entry.dir) {
      path += '/' + entry.name;
    }
    this.clipboard.copy(path);
  }

  @HostListener('window:resize', ['$event'])
  onResize() {
    this.isMobile = window.innerWidth < 768;
  }

  public getResult() {
    return this.result.asObservable();
  }

  closeDialog() {
    this.result.next(undefined);
    this.resultDir.emit(undefined);
    this.result.complete();
    this.visibleChange.emit(false);
  }

  confirm() {
    let path = this.routeStack.items.join('/');
    if (path.startsWith('/')) {
      path = path.substring(1);
    }
    this.result.next(path);
    this.resultDir.emit(path);
    this.result.complete();
    this.visibleChange.emit(false);
  }

  private loadChildren(dir: string) {
    this.ioService.ls(dir, this.showFiles).subscribe({
      next: (entries) => {
        this.entries = entries || [];
      },
      error: (err) => {
        this.routeStack.pop();
        this.msgService.error("Failed to load children", err.error.message)
      }
    })
  }

}
