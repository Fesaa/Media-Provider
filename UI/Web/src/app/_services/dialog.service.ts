import {
  Injectable,
  ComponentRef,
  ApplicationRef,
  ViewContainerRef
} from '@angular/core';
import {DialogComponent} from "../shared/_component/dialog/dialog.component";
import {DirectorySelectorComponent} from "../shared/_component/directory-selector/directory-selector.component";

@Injectable({
  providedIn: 'root'
})
export class DialogService {

  public viewContainerRef: ViewContainerRef | undefined;

  constructor(
    private appRef: ApplicationRef,
  ) {}

  openDirBrowser(root: string): Promise<string| undefined> {
    const component = this.viewContainerRef!.createComponent(DirectorySelectorComponent)
    component.instance.root = root;
    component.instance.filter = true;
    component.instance.copy = false;

    return new Promise<string | undefined>((resolve, reject) => {
      component.instance.getResult().subscribe(result => {
        this.closeDialog(component);
        resolve(result);
      });
    });
  }

  openDialog(text: string): Promise<boolean> {
    const component = this.viewContainerRef!.createComponent(DialogComponent)
    component.instance.text = text;

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
