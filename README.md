# ratelimit
```bash
./ratelimit -h
Usage of ./ratelimit:
./ratelimit --rate <N> --inflight <P> <command...>
<command...>: the command to launch, '{}' is replaced with string from stdin
  -inflight int
        maximum number of concurrently running commands (default 1)
  -rate int
        maximum number of command launches per second (default 1)

```


## RUN

```bash
git clone https://github.com/DenisMV-B/ratelimit.git && cd ratelimit && make
```

## Описание реализации

1) Контролирование количества одновременно работающих goroutine выполняется с помощью семафора, который реализован в виде буферизированного канала вместимостью inflight

2) Контролирование частоты запуска команд выполняется с помощью буферизированного канала и time.Tick https://gobyexample.com/rate-limiting

3) Акцент реализации сделан на простоту, используются примитивные механизмы. Контролирование одновременно работающих goroutine можно так же реализовать в виде worker pool (https://gobyexample.com/worker-pools). Т.е. создать [inflight] worker'ов, положить задачи (под задачей понимается команда с аргументами, которую необходимо выполнить) в буферизированный канал, затем worker'ы будут конкурентно забирать задачи из канала и выполнять их. На мой взгляд, эта реализация более многословная, чем моя. Готов реализовать другим образом, если требуется.

4) Немного не понял output в условии:   
```bash
$ for i in {1..60} ; do echo $i ; done | ./ratelimit --rate 1 --inflight 15 echo {}
Эта команда должна отработать за ~4с.

$ (echo 1 ; sleep 3 ; echo 2 ; echo 3) | ./ratelimit --rate 1 --inflight 2 echo {}
Эта команда должна отработать за ~4с.
```
Если rate == 1 => разница во времени между запусками команды  - не более 1 секунды, значит 60 команд отработают ~59 сек  
Похожих результатов мне удалось достигнуть, если распределять это время между [inflight] goroutine'ами (в коде есть закомментированные участки по этому поводу (inflight*rate))

Возможно я неправильно понял условие, если требуется, то скорректирую работу утилиты по уточнениям
