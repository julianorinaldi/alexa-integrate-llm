# 📂 Volume Data

Este diretório (`volume/data`) é utilizado para mapear as persistências locais do container Docker para o seu sistema de arquivos local.

Principalmente, o banco de dados SQLite (`alexa.db`) do projeto será criado e mantido aqui. 
Isto garante que, mesmo que o container seja destruído, atualizado ou recriado, todos os seus dados (como usuários do dashboard e acessos das skills) permanecerão salvos com segurança de forma contínua e fácil de acessar.

⚠️ **IMPORTANTE:**
- Não exclua o arquivo `alexa.db` daqui, a menos que queira resetar o banco de dados completamente.
- Este diretório será compactado ao executar o comando de backup (`make backup`).
