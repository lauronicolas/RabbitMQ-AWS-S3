Este sistema foi desenvolvido com o intuito de entender o funcionamento de integração entre um aplicação GO com o AWS S3.

Junto a ele foi integrado um sistema de mensageria para criação de logs de controle utilizando o RabbitMQ.

Antes de executar a aplicação principal, é aconselhavel executar a aplicação "generator" que irá gerar 1 Milhao de arquivos simples em .txt para serem mandados para o S3.

A quantidade de arquivos acaba sendo pequena, visto que a aplicação utiliza de 100 Go Routines para fazer o upload de forma assincrona.
