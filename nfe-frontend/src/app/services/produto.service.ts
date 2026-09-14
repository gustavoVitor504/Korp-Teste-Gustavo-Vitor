import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import { Produto } from '../models/produto.model';

@Injectable({ providedIn: 'root' })
export class ProdutoService { // recebe URL com base no environment
  private readonly baseUrl = `${environment.estoqueApiUrl}/produtos`;

  constructor(private http: HttpClient) {} // injeção de dependências

  listar(): Observable<Produto[]> { // observable recebe os valores e preenche o array do tipo Produto[]
    return this.http.get<Produto[]>(this.baseUrl);
  }

  criar(produto: Produto): Observable<Produto> { // passa produto no endpoint de criar e recebe de volta o produto
    return this.http.post<Produto>(this.baseUrl, produto);
  }
}
