import { ComponentFixture, TestBed } from '@angular/core/testing';

import { RunningInfoComponent } from './running-info.component';

describe('RunningInfoComponent', () => {
  let component: RunningInfoComponent;
  let fixture: ComponentFixture<RunningInfoComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [RunningInfoComponent]
    })
    .compileComponents();

    fixture = TestBed.createComponent(RunningInfoComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
