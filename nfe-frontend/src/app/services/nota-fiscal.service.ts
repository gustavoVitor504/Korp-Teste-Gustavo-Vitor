import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { environment } from '../../environments/environment';
import { CriarNotaFiscalPayload, NotaFiscal } from '../models/nota-fiscal.model';

@Injectable({ providedIn: 'root' })
export class NotaFiscalService { // recebe a URL com base no environment
  private readonly baseUrl = `${environment.faturamentoApiUrl}/notas-fiscais`;

  constructor(private http: HttpClient) {} // injeção de dependência

  listar(): Observable<NotaFiscal[]> { // observable alimentado com os dados da requisição
    return this.http
      .get<NotaFiscal[]>(this.baseUrl) // tipagem de retorno esperado
      .pipe(map((notas) => notas.map((n) => this.normalizar(n))));
  }

  criar(payload: CriarNotaFiscalPayload): Observable<NotaFiscal> { // recebe o payload criado
    return this.http
      .post<NotaFiscal>(this.baseUrl, payload) // tipa como NotaFiscal porque espera receber de volta uma nota
      .pipe(map((n) => this.normalizar(n)));
  }

  imprimir(id: number): Observable<NotaFiscal> { // recebe o ID
    return this.http
      .post<NotaFiscal>(`${this.baseUrl}/${id}/imprimir`, {}) // faz requisição de imprimir com o ID fornecido
      .pipe(map((n) => this.normalizar(n)));
  }

  private normalizar(nota: NotaFiscal): NotaFiscal {
    return { ...nota, itens: nota.itens ?? [] };
  } // componente trabalha esperando um [] e por padrão o GO tranforma um [] vazio em undefined
}
