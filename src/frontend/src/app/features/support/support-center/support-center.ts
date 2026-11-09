import { Component, OnInit, inject } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { TranslateModule } from '@ngx-translate/core';
import { MatIconModule } from '@angular/material/icon';
import {
  CreateSupportCaseRequest,
  SupportCase,
  SupportCaseStatus,
} from '../../../core/models/support-case.model';
import { SupportService } from '../../../core/services/support.service';

@Component({
  selector: 'app-support-center',
  standalone: true,
  imports: [FormsModule, MatIconModule, TranslateModule],
  templateUrl: './support-center.html',
  styleUrl: './support-center.scss',
})
export class SupportCenter implements OnInit {
  private readonly supportService = inject(SupportService);

  cases: SupportCase[] = [];
  selectedCase: SupportCase | null = null;
  newCase: CreateSupportCaseRequest = { subject: '', description: '' };
  replyBody = '';
  loadingCases = false;
  loadingCase = false;
  submittingCase = false;
  submittingReply = false;
  error = '';
  notice = '';

  ngOnInit(): void {
    this.loadCases();
  }

  loadCases(selectCaseId?: number): void {
    this.loadingCases = true;
    this.supportService.listCases().subscribe({
      next: ({ cases }) => {
        this.cases = cases;
        this.loadingCases = false;
        if (selectCaseId !== undefined) {
          const created = cases.find((supportCase) => supportCase.id === selectCaseId);
          if (created) this.openCase(created);
        }
      },
      error: (error: HttpErrorResponse) => {
        this.loadingCases = false;
        this.showError(error);
      },
    });
  }

  openCase(supportCase: SupportCase): void {
    this.error = '';
    this.notice = '';
    this.loadingCase = true;
    this.selectedCase = supportCase;
    this.supportService.getCase(supportCase.id).subscribe({
      next: (loadedCase) => {
        this.selectedCase = loadedCase;
        this.loadingCase = false;
      },
      error: (error: HttpErrorResponse) => {
        this.loadingCase = false;
        this.showError(error);
      },
    });
  }

  createCase(): void {
    const request = {
      subject: this.newCase.subject.trim(),
      description: this.newCase.description.trim(),
    };
    if (!request.subject || !request.description) return;

    this.error = '';
    this.notice = '';
    this.submittingCase = true;
    this.supportService.createCase(request).subscribe({
      next: (created) => {
        this.submittingCase = false;
        this.newCase = { subject: '', description: '' };
        this.notice = 'SUPPORT_CENTER.CASE_CREATED';
        this.loadCases(created.id);
      },
      error: (error: HttpErrorResponse) => {
        this.submittingCase = false;
        this.showError(error);
      },
    });
  }

  addReply(): void {
    const supportCase = this.selectedCase;
    const body = this.replyBody.trim();
    if (!supportCase || !body) return;

    this.error = '';
    this.notice = '';
    this.submittingReply = true;
    this.supportService.addMessage(supportCase.id, body).subscribe({
      next: (message) => {
        this.selectedCase = {
          ...supportCase,
          messages: [...(supportCase.messages ?? []), message],
          updated_at: message.created_at,
        };
        this.cases = this.cases.map((item) =>
          item.id === supportCase.id ? { ...item, updated_at: message.created_at } : item
        );
        this.replyBody = '';
        this.submittingReply = false;
      },
      error: (error: HttpErrorResponse) => {
        this.submittingReply = false;
        this.showError(error);
      },
    });
  }

  statusTranslation(status: SupportCaseStatus): string {
    return `SUPPORT_CENTER.STATUS.${status.toUpperCase()}`;
  }

  formatDate(date: string): string {
    return new Date(date).toLocaleString();
  }

  private showError(error: HttpErrorResponse): void {
    this.error = error.error?.error ?? 'SUPPORT_CENTER.ERROR';
  }
}
