; Advent of code 2020, Day 8
(defun aoc8 () 
    (def input (split (readFile "samples/aoc8.txt") "\n"))
    (def acc 0)
    (def seenInstructions (list))
    (def i 0)
    (def finalAcc null)

    (while (= finalAcc null) 
        (def instrLine (nth i input))
        (def instr (substr instrLine 0 3))
        (def sign (nth 4 instrLine))
        (def argMag (strToNum (substr instrLine 5 10)))
        (def arg (if (= sign "-") (* -1 argMag) argMag))
        
        (if (contains seenInstructions i)
            (def finalAcc acc))

        (def seenInstructions (append i seenInstructions))

        (if (= instr "nop")
            (def i (+ i 1)))
        
        (if (= instr "jmp")
            (def i (+ i arg)))

        (if (= instr "acc")
            ((def acc (+ acc arg))
            (def i (+ i 1))))
    )

    (finalAcc)
)

(defun isPrime (n)
    (def prime true)
    (if (= n 1)
        false
        ((def i 2)
        (while (and prime (<= i (sqrt n)))
            (if (= 0 (mod n i))(def prime false))
            (def i (+ i 1)))))
    (prime)
)

; Project euler 7
(defun euler7 ()
    (def primeCount 0)
    (def i 2)
    (while (not (= primeCount 10001))
        (if (isPrime i)(
            (def primeCount (+ 1 primeCount))))
        (def i (+ i 1))
    )
    (- i 1)
)

(defun main ()
    (input)
)