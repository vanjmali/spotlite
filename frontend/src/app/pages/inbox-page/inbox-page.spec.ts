import { ComponentFixture, TestBed } from '@angular/core/testing';

import { InboxPage } from './inbox-page';

describe('InboxPage', () => {
  let component: InboxPage;
  let fixture: ComponentFixture<InboxPage>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [InboxPage],
    }).compileComponents();

    fixture = TestBed.createComponent(InboxPage);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
