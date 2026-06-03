# Canales de Comunicación (Channels)
Los canales son los conductos tipados a través de los cuales puedes enviar y recibir valores con el operador de canal, <-.
Se inicializan usando make: ch := make(chan int).
Por defecto, los envíos y recepciones se bloquean hasta que el otro lado esté listo, lo que permite sincronizar goroutines de forma nativa sin bloqueos explícitos ni variables de condición.