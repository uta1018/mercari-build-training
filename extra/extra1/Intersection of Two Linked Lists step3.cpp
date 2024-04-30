#include <bits/stdc++.h>
using namespace std;

/**
 * Definition for singly-linked list.
 * struct ListNode {
 *     int val;
 *     ListNode *next;
 *     ListNode(int x) : val(x), next(NULL) {}
 * };
 */
class Solution {
public:
    ListNode *getIntersectionNode(ListNode *headA, ListNode *headB) {
        // headAのtailからheadBにポインタを張る
        ListNode *tail = headA;
        while (tail->next != nullptr) {
            tail = tail->next;
        }
        tail->next = headB;

        // ループがあるか調べる　Floyd's Linked List Cycle Finding Algorithm
        ListNode *fast = headA, *slow = headA;
        while(fast && fast->next) {
            fast = fast->next->next;
            slow = slow->next;
            if(fast == slow) break;
        }
        if(!fast || !fast->next) {
            tail->next = nullptr;
            return nullptr;
        }
        
        // ループの開始点を求める
        slow = headA;
        while(slow != fast) {
            slow = slow->next;
            fast = fast->next;
        }
        tail->next = nullptr;
        return fast;
    }
};

