import os
import psutil

print("Buscando processos com nome 'proxy_universal_out'...")
for p in psutil.process_iter(['pid', 'name']):
    if p.info['name'] == 'proxy_universal_out':
        print(f"Encontrado! PID: {p.info['pid']}")
        try:
            with open(f"/proc/{p.info['pid']}/environ", "rb") as f:
                env = f.read()
                print("Extraído do environ:")
                parts = env.split(b'\0')
                for part in parts:
                    if b"CROM_TENANT_SEED" in part:
                        print("=> VULNERABILIDADE DETECTADA:")
                        print("  ", part.decode(errors='ignore'))
        except Exception as e:
            print("Erro ao ler:", e)
