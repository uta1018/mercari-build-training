#include <bits/stdc++.h>
using namespace std;

class Solution {
public:
    bool wordPattern(string pattern, string s) {
        vector<string> v; // 分割した文字列を格納するvector
        string str;
        stringstream ss{s}; // 入出力可能なsstreamに変換

        // Step1
        while ( getline(ss, str, ' ') ){     // スペース（' '）で区切って，格納
            v.push_back(str);
        }

        // Step2
        // パターン p の各文字が、s のどの部分に対応するかを管理するmap
        map<char, string> patternToWord;  
        // s の分割された各文字列が、パターン p のどの文字に対応するかを管理するmap
        map<string, char> wordToPattern; 

        // 文字列の長さが異なるとき
        if(v.size() != pattern.size()) return false;

        for(int i=0;i<pattern.size();i++) {
            // sの分割された文字列もパターンpの文字も割り当てがないとき
            if(patternToWord.find(pattern.at(i)) == patternToWord.end() && wordToPattern.find(v[i]) ==  wordToPattern.end()) {
                patternToWord[pattern.at(i)] = v[i];
                wordToPattern[v[i]] = pattern.at(i);
            } 
            // 割り当てがあって、それが一致するとき
            else if(patternToWord[pattern.at(i)] == v[i]) continue; 
            else return false;
        }
        return true;
    }
};
