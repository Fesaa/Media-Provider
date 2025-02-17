import {ApplicationRef, ComponentRef, Injectable, ViewContainerRef} from '@angular/core';
import {DialogComponent} from "../shared/_component/dialog/dialog.component";
import {DirectorySelectorComponent} from "../shared/_component/directory-selector/directory-selector.component";
import {TranslocoService} from "@jsverse/transloco";

@Injectable({
  providedIn: 'root'
})
export class DialogService {

  public viewContainerRef: ViewContainerRef | undefined;

  constructor(
    private appRef: ApplicationRef,
    private transLoco: TranslocoService,
  ) {
  }

  openDirBrowser(
    root: string,
    options: Partial<{ create: boolean; copy: boolean; filter: boolean; showFiles: boolean, width: string }> = {}
  ): Promise<string | undefined> {
    const defaultOptions = {create: false, copy: false, filter: true, showFiles: false, width: '50vw'};
    const finalOptions = {...defaultOptions, ...options};

    const component = this.viewContainerRef!.createComponent(DirectorySelectorComponent)
    component.instance.root = root;
    component.instance.filter = finalOptions.filter;
    component.instance.copy = finalOptions.copy;
    component.instance.create = finalOptions.create;
    component.instance.showFiles = finalOptions.showFiles;
    component.instance.customWidth = finalOptions.width

    return new Promise<string | undefined>((resolve, reject) => {
      component.instance.getResult().subscribe(result => {
        this.closeDialog(component);
        resolve(result);
      });
    });
  }

  openDialog(text: string, textArgs?: any, header: string = "Confirm", headerArgs?: any): Promise<boolean> {
    const component = this.viewContainerRef!.createComponent(DialogComponent)
    component.instance.text = this.transLoco.translate(text, textArgs);
    component.instance.header = this.transLoco.translate(header, headerArgs);

    return new Promise<boolean>((resolve, reject) => {
      component.instance.getResult().subscribe(result => {
        this.closeDialog(component);
        resolve(result);
      });
    });
  }

  closeDialog(componentRef: ComponentRef<any>) {
    this.appRef.detachView(componentRef.hostView);
    componentRef.destroy();
  }
}
