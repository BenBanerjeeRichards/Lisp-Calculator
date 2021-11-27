; Advent of code 2020, Day 8
(defun aoc8 () 
    (def input (split (readFile "aoc8.txt") "\n"))
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