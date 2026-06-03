# Fundamentos de Concurrencia: Goroutines
Una goroutine es un hilo de ejecución ligero administrado por el runtime de Go.
Sintaxis: go f(x, y, z) arranca una nueva goroutine que ejecuta f.
Las goroutines comparten el mismo espacio de direcciones, por lo que el acceso a la memoria compartida debe ser sincronizado utilizando canales o el paquete sync de la biblioteca estándar.