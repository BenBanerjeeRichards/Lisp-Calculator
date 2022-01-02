(import "calc/stdlib.lisp" "lib")


(defun assertEq (a b message)
    (if (not (= a b))
        (printLn (concatAll (list "FAILED (" message ")" " Expected=" a " Actual=" b))))
)

(defun testConcatAll () 
    (assertEq ("hello world") (concatAll (list "hello" " " "world")) "testConcatAll1")
    (assertEq ("hellotruenull34") (concatAll (list "hello" true null 34)) "testConcatAll2")
    (assertEq ("hello(1 false \"world\")") (concatAll (list "hello" (list 1 false "world"))) "testConcatAll3")
)

(defun testAppend () 
    (assertEq (list 1 2 3) (append 3 (list 1 2)) "append1")
    (assertEq (list 3) (append 3 (list)) "append2")
)

(defun testMap () 
    (def double (lambda (x) (* x 2)))
    (assertEq (list 2 4 6) (map (list 1 2 3) double) "map1")
    (assertEq (list) (map (list) double) "map2")
)

(defun testFilter () 
    (assertEq (list 1 2 3) (filter (list 1 2 3 4 5 6 7 8 9) (lambda (x) (<= x 3)))"filter1")
)


(defun testReduce () 
    (assertEq 6 (reduce (list 1 2 3) 0 (lambda (a b) (+ a b))) "reduce1")
    (assertEq 6 (reduce (list 1 2 3) 1 (lambda (a b) (* a b))) "reduce2")
)

(defun testRange ()
    (assertEq (list 1 2 3 4 5) (range 1 6 1) "range1")
    (assertEq (list 1 3 5) (range 1 6 2) "range2")
    (assertEq (list 5 4 3 2 1) (range 5 0 -1) "range3")
)

(defun testSubstr () 
    (assertEq "abcd" (substr "abcd" 0 4) "substr1")
    (assertEq "abc" (substr "abcd" 0 3) "substr2")
    (assertEq "ab" (substr "abcd" 0 2) "substr3")
    (assertEq "bc" (substr "abcd" 1 3) "substr4")
    (assertEq "" (substr "abcd" 3 0) "substr5")
    (assertEq "" (substr "abcd" 0 0) "substr6")
    (assertEq "a" (substr "abcd" 0 1) "substr7")
)

(defun testSplit () 
    (assertEq (list "the" "quick" "brown" "fox") (split "the quick brown fox" " ") "split1")
    (assertEq (list "the" "quick" "brown" "fox" "") (split "the quick brown fox " " ") "split2")
    (assertEq (list "" " mat was sat on by " " cat by " " lamp") (split "the mat was sat on by the cat by the lamp" "the") "split3")
    (assertEq (list "" " mat was sat on by " " cat by " " lamp " "") (split "the mat was sat on by the cat by the lamp the" "the") "split4")
    (assertEq (list "the quick brown fox") (split "the quick brown fox" "@") "split5")
    (assertEq (list "the quick brown fox") (split "the quick brown fox" "deliminator") "split6")
    (assertEq (list "the quick brown fox") (split "the quick brown fox" "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") "split7")
)

(defun testCharToNum ()
    (assertEq 5 (charToNum "5") "charToNum1")
    (assertEq 0 (charToNum "0") "charToNum2")
    (assertEq 9 (charToNum "9") "charToNum3")
)

(defun testStrToNum ()
    (assertEq 123 (strToNum "123") "strToNum1")
    (assertEq 0 (strToNum "0") "strToNum2")
    (assertEq 123.456 (strToNum "123.456") "strToNum3")
    (assertEq 0.5 (strToNum "0.5") "strToNum4")
    (assertEq  5.0 (strToNum "5.0") "strToNum5")
)

(defun testContains ()
    (assertEq true (contains (list 1 2 3) 3) "contains1")
    (assertEq true (contains (list 1 2 3) 1) "contains2")
    (assertEq false (contains (list 1 2 3) 5) "contains2")
    (assertEq false (contains (list) 5) "contains4")
)

(defun testRound ()
    (assertEq 10 (round 10) "round1")
    (assertEq 10 (round 10.3) "round2")
    (assertEq 10 (round 10.49) "round3")
    (assertEq 11 (round 10.5) "round4")
    (assertEq 11 (round 10.7) "round5")
    (assertEq 11 (round 11) "round6")
)

(defun test () 
    (testAppend)
    (testMap)
    (testConcatAll)
    (testFilter)
    (testReduce)
    (testRange)
    (testSubstr)
    (testSplit)
    (testCharToNum)
    (testStrToNum)
    (testContains)
    (testRound)
)

(defun main (args)
    (test)
)

