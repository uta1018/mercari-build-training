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
        map<ListNode*, bool> Map;
        // listAのノードを記録する
        while (headA != nullptr) {
            Map[headA] = true;
            headA = headA->next;
        }

        // 交差点を探す
        while (headB != nullptr) {
            if(Map.find(headB) != Map.end()) return headB;
            headB = headB->next;
        }

        return nullptr;
    }
};

