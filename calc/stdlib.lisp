(defun assertEq (a b message)
    (if (not (= a b))
        (printLn (concatAll (list "FAILED (" message ")" " Expected=" a " Actual=" b))))
)

(defun printLn (x) 
    (print x)
    (print "\n")
)

(defun concatAll (x) 
    (reduce x "" 
        (lambda (a b) (concat a b)))
)

(defun append (x xs)
    (insert (length xs) x xs)
)


(defun map (items f)
    (def i 0)
    (def resultList (list))
    (while (< i (length items))
        (def resultList (append 
            (funcall f (nth i items))
            resultList
        ))
        (def i (+ i 1))
    )
    (resultList)
)

(defun filter (items f)
    (def i 0)
    (def resultList (list))
    (while (< i (length items))
        (def val (nth i items))
        (if (funcall f val)
            (def resultList (append val resultList))
        )
        (def i (+ i 1))
    )
    (resultList)
)

(defun reduce (items initial f)
    (def i 0)
    (def result (if (= null initial) (nth 0 items) (initial)))
    (while (< i (length items))
        (def val (nth i items))
        (def result (funcall f result val))
        (def i (+ i 1))
    )
    (result)
)

(defun range (from to increment)
    (if 
        (and (= increment 0) (not (= from to)))
        (panic "range will generate infinite list "))

    (def rangeSign (< from to))
    (def incrementSign (> increment 0))
    (if 
        (not (= rangeSign incrementSign))
        (panic "range will generate infinite list "))

    (def i from)
    (def result (list))

    (while (if (> increment 0) (< i to) (> i to))
        (def result (append i result))
        (def i (+ i increment))
    )
    (result)
)

(defun substr (string from to)
    (def to 
    (if (>= to (length string))
        (length string)
        to))

    (if (> from to) ("")
        ((def result "")
            (def i from)
            (while (< i to)
                (def result (concat result (nth i string)))
                (def i (+ i 1)))
            (result)))
)

(defun split (string delim)
    (def result (list))
    (def acc "")
    (def i 0)
    (def n (length delim))

    (while (< i (length string))
        (def subDelim (substr string i (+ i n)))
        (if (= subDelim delim)
            (
                (def result (append acc result))
                (def acc "")
                (def i (+ i n))
            )
            (
                (def acc (concat acc (nth i string)))
                (def i (+ i 1))
            )))

    (def result (append acc result))
    (result)
)

(defun first (items) (nth 0 items))
(defun second (items) (nth 1 items))

(defun charToNum (char) 
    (def code (ord char))
    (if (or (> code 58) (< code 48))
        (panic (concatAll (list "charToNum - `" char "` is not a valid digit")))
        (- code 48)) 
)

; strToNum converts a string to a number
; Supports either integers (1231) or floating point (234.123)
; Only these two basic formats supported 
(defun strToNum (string)
    (def parts (split string "."))
    (if (> (length parts) 2)
        (panic (concat "strToNum - invalid input" string)))
    
    (def whole 0)
    (def i 0)
    (def wholeStr (first parts))

    (while (< i (length wholeStr))
        (def digit 
            (charToNum 
                (nth (- (- (length wholeStr) i) 1) wholeStr)))

        (def whole (+ whole (
            (* digit (^ 10 i)))))
        (def i (+ i 1))
    )

    (if (= (length parts) 2)(
        (def i 0)
        (def fract 0)
        (def fractStr (second parts))
        (while (< i (length fractStr))
            (def digit (charToNum (nth i fractStr)))
            (def fi (+ i 1))
            (def fract (+ fract (
                (* digit (/ 1 (^ 10 fi)))
            )))
            (def i (+ i 1))
        ) 
        (+ whole fract))
        (whole))
)

(defun contains (xs x)
    (> (length (filter xs (lambda (a) (= a x)))) 0)
)
