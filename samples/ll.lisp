(import "calc/stdlib.lisp")

(defstruct linkedlist_node value next)
(defstruct linkedlist head)

(defun ll_append (llist value)
    (def new_node (struct linkedlist_node (value value)))

    (if (= llist:head null)
        (def llist:head new_node)
        (
            (def current_node llist:head)
            (while (not (= null (current_node:next)))
                (def current_node (current_node:next)))

            (def current_node:next new_node)
            (llist)
        )   
    )
)

(defun ll_to_list (llist)
    (def l (list))
    (def node llist:head)
    (while (not (= null node))
        (def l (append (node:value) l))
        (def node (node:next)))
    (l)
)
(defun ll_remove (llist value_to_remove)
    (def node llist:head)
    (if (= node:value value_to_remove) (
        (return (struct linkedlist (head node:next)))))

    (def prev llist:head)
    (def curr prev:next)

    (while  (not (= null curr))
        (if (= curr:value value_to_remove) (
            (def prev:next curr:next)
            (return (struct linkedlist (head node)))
        ))
        (def prev curr)
        (def curr curr:next))

    (llist)
)

(defun ll_exists (llist value)
    (def node llist:head)
    (while (not (= node null))
        (if (= value node:value)(return true))
        (def node node:next)
    )
    (return false)
)


(def ll (struct linkedlist))
(ll_append ll 23)
(ll_append ll 24)
(ll_append ll 26)

(ll_to_list ll)

(ll_exists ll 24)