package com.rvcs.android

import android.os.Bundle
import android.util.Log
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity

class MainActivityTest : AppCompatActivity() {
    companion object {
        private const val TAG = "MainActivityTest"
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        try {
            Log.d(TAG, "onCreate: started")
            setContentView(R.layout.activity_main)

            val textView = TextView(this)
            textView.text = "Test App - This is working!"
            textView.textSize = 24f
            textView.textAlignment = TextView.TEXT_ALIGNMENT_CENTER

            setContentView(textView)

            Log.d(TAG, "onCreate: completed successfully")
        } catch (e: Exception) {
            Log.e(TAG, "onCreate: Exception", e)
            e.printStackTrace()
        }
    }
}
