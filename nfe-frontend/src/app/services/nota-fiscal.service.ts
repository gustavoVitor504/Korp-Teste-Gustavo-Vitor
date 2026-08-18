import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { environment } from '../../environments/environment';
import { CriarNotaFiscalPayload, NotaFiscal } from '../models/nota-fiscal.model';

@Injectable({ providedIn: 'root' })
export class NotaFiscalService {
  private readonly baseUrl = `${environment.faturamentoApiUrl}/notas-fiscais`;

  constructor(private http: HttpClient) {}

  listar(): Observable<NotaFiscal[]> {
    return this.http
      .get<NotaFiscal[]>(this.baseUrl)
      .pipe(map((notas) => notas.map((n) => this.normalizar(n))));
  }

  criar(payload: CriarNotaFiscalPayload): Observable<NotaFiscal> {
    return this.http
      .post<NotaFiscal>(this.baseUrl, payload)
      .pipe(map((n) => this.normalizar(n)));
  }

  imprimir(id: number): Observable<NotaFiscal> {
    return this.http
      .post<NotaFiscal>(`${this.baseUrl}/${id}/imprimir`, {})
      .pipe(map((n) => this.normalizar(n)));
  }

  private normalizar(nota: NotaFiscal): NotaFiscal {
    return { ...nota, itens: nota.itens ?? [] };
  }
}
