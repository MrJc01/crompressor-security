# Interagindo com o Docker e Outros Agentes via VFS

Um dos focos de extrema proteção em ambientes descentralizados onde hospedas containers e infraestrutura cloud-native como o Docker, é assegurar que o disco não possa ser corrompido, adulterado silenciosamente ou que exfiltre conhecimento em forma de Data-Lakes abertos.

## O Conceito "Engolir o Docker"

O **Crompressor** soluciona isso engolindo a camada de hardware Virtual File System (VFS/FUSE), tornando qualquer gravação "opaca" ao sistema hospedeiro (Node SO original) e totalmente rastreável reversível pela rede Gossip via Árvores de Merkle.

### Aplicação de Mount FUSE nativa:

No projeto base do Crompressor, já existe uma camada de compilação em `pkg/cromlib/vfs/fuse_server.go` e `internal/vfs/fuse.go`.

Nós utilizamos uma tática de hiper-isolamento:
Ao rodar instâncias ou serviços Docker Críticos, redirecionamos todos os mappings do de volume de disco para a abstração do Fuse do Crompressor:

1. `docker run -v /vols_docker/app_critico:/app ...`  onde o HD físico em `/vols_docker` é interceptado pelo FileSystem Crompressor.
2. Cada Gravação (IOPS) feita pelo Engine (ex: gravar um arquivo temporário json no disco) dispara um Hook.
3. O VFS do Crompressor pega a string binária sendo escrita, divide (Chunker CDC/FastCDC via `internal/chunker/`), identifica duplicidades no espaço latente da rede inteira e salva com compressão altíssima no bloco.
4. Qualquer tentativa de leitura direta por um vírus de Host nos blocos armazenados sob `/vols_docker` nativamente resulta na leitura de estilhaços de entropia vazia (Diferente da criptografia padrão LUKS, o Crompressor esmaga os dados por referência de codebook, extraindo a informação sensível do disco em 95%). Só a chave correta no daemon do `fuse_server.go` remolda ao bloco Unix nativo para o Docker ler de volta com o cache (CromFS/VFS Cache).

### Rastreabilidade e Árvore de Eventos

> [!WARNING] 
> Todo bloco gravado no CROMFS-FUSE emite uma folha na `internal/merkle`. 
> Se um atacante usar uma tática "Supply Chain" rodando dentro de seu container, criptografando tudo com um *Ransomware* silencioso ou apagando as tables:
> Você reverte no tempo usando os metadados das últimas folhas Hash Válidas, ignorando instantaneamente os deltas da infecção. Tempo de reparo: Near Zero.
